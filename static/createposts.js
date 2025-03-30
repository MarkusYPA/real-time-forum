import { handleDislike, handleLike, openAndSendReply, openReplies } from "./posts.js";
import { feed } from "./realtime.js";

export function formatDate(isoString) {
    const date = new Date(isoString);

    // Get Finland's time zone offset dynamically
    const options = { 
        timeZone: "Europe/Helsinki", 
        day: "2-digit", 
        month: "2-digit", 
        year: "numeric", 
        hour: "2-digit", 
        minute: "2-digit", 
        hour12: false 
    };

    return new Intl.DateTimeFormat("fi-FI", options).format(date).replace("klo ","");
}

// Function to add a post to the page
export function addPostToFeed(post) {
    const newPost = document.createElement('div');
    newPost.className = 'post';
    const formattedID = `postid${post.id}`;
    newPost.id = formattedID;

    const postItems = document.createElement('div');
    const rowTitle = document.createElement('div');
    const title = document.createElement('div');
    const rowAuthorDate = document.createElement('div');
    const author = document.createElement('span');
    const date = document.createElement('span');
    const content = document.createElement('div');
    const rowLikes = document.createElement('div');
    const likesThumb = document.createElement('span');
    const likesText = document.createElement('span');
    const dislikesThumb = document.createElement('span');
    const dislikesText = document.createElement('span');
    const rowAddRepy = document.createElement('div');
    const addReplySymbol = document.createElement('span');
    const addReplyText = document.createElement('span');
    const repliesInfo = document.createElement('span');
    const rowBottom = document.createElement('div');
    const addReplyDiv = document.createElement('div');
    const replyDiv = document.createElement('div');

    postItems.classList.add('post-items');
    rowTitle.classList.add('row');
    title.classList.add('post-title');
    rowAuthorDate.classList.add('row')
    author.classList.add('post-author');
    date.classList.add('post-date');
    content.classList.add('post-content');
    rowLikes.classList.add('row', 'post-reactions');
    likesThumb.classList.add('material-symbols-outlined', 'likes', 'likes-tumb');
    likesText.classList.add('post-likes');
    dislikesThumb.classList.add('material-symbols-outlined', 'likes', 'dislikes-tumb');
    dislikesText.classList.add('post-dislikes');
    rowAddRepy.classList.add('row', 'post-addition');
    addReplySymbol.classList.add('material-symbols-outlined', 'likes');
    addReplyText.classList.add('post-addreply');
    repliesInfo.classList.add('post-replies');
    repliesInfo.id = `post-${post.id}`
    rowBottom.classList.add('row');
    addReplyDiv.classList.add('add-reply');
    replyDiv.classList.add('replies');

    title.textContent = post.title;
    author.textContent = post.user.username;
    date.textContent = formatDate(post.created_at);
    content.textContent = post.description;
    likesThumb.textContent = "thumb_up";
    likesText.textContent = post.number_of_likes;
    dislikesThumb.textContent = "thumb_down";
    dislikesText.textContent = post.number_of_dislikes;
    addReplySymbol.textContent = "chat_bubble"
    addReplyText.textContent = "add reply"
    repliesInfo.textContent = post.repliesCount + " replies";

    if (post.liked) likesThumb.style.color = "green"
    if (post.disliked) dislikesThumb.style.color = "red"

    likesThumb.addEventListener("click", () => handleLike(post.id, "post"));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id, "post"));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, post.id, "post"))

    if (post.repliesCount > 0) {
        repliesInfo.classList.add('clickable');
        repliesInfo.addEventListener("click", () => openReplies(post.id, "post", formattedID, replyDiv));
    }

    rowTitle.appendChild(title);
    rowAuthorDate.appendChild(author);
    rowAuthorDate.appendChild(date);
    postItems.appendChild(rowAuthorDate);
    postItems.appendChild(rowTitle);
    postItems.appendChild(content);
    rowLikes.appendChild(likesThumb);
    rowLikes.appendChild(likesText);
    rowLikes.appendChild(dislikesThumb);
    rowLikes.appendChild(dislikesText);
    rowBottom.appendChild(rowLikes);
    rowAddRepy.appendChild(addReplySymbol);
    rowAddRepy.appendChild(addReplyText);
    rowBottom.appendChild(rowAddRepy);
    rowBottom.appendChild(repliesInfo);

    post.categories.forEach(cat => {
        const category = document.createElement('span');
        category.classList.add('post-categories');
        category.textContent = cat.name;
        rowBottom.appendChild(category);
    });

    postItems.appendChild(rowBottom);
    postItems.appendChild(addReplyDiv);
    newPost.appendChild(postItems);
    newPost.appendChild(replyDiv);

    content.style.display = "none";
    rowBottom.style.display = "none";
    replyDiv.style.display = "none";
    title.addEventListener("click", () => {
        content.style.display == "none" ? content.style.display = "block" : content.style.display = "none";
        rowBottom.style.display == "none" ? rowBottom.style.display = "flex" : rowBottom.style.display = "none";
        replyDiv.style.display == "none" ? replyDiv.style.display = "block" : replyDiv.style.display = "none";
        // remove possible reply input
        const existingReplyInput = newPost.querySelector('.reply-container');
        if (existingReplyInput) existingReplyInput.remove();
    })

    feed.prepend(newPost);
}

export function addReplyToParent(parentFormattedID, comment, numberOfRepliesForParent) {
    const parent = document.getElementById(parentFormattedID);

    const newReply = document.createElement('div');
    newReply.className = 'reply';
    const formattedID = `replyid${comment.id}`;
    newReply.id = formattedID;

    const replyItems = document.createElement('div');
    const rowTitle = document.createElement('div');
    const rowAuthorDate = document.createElement('div');
    const author = document.createElement('span');
    const date = document.createElement('span');
    const content = document.createElement('div');
    const rowLikes = document.createElement('div');
    const likesThumb = document.createElement('span');
    const likesText = document.createElement('span');
    const dislikesThumb = document.createElement('span');
    const dislikesText = document.createElement('span');
    const rowAddRepy = document.createElement('div');
    const addReplySymbol = document.createElement('span');
    const addReplyText = document.createElement('span');
    const repliesInfo = document.createElement('span');
    const rowBottom = document.createElement('div');
    const addReplyDiv = document.createElement('div');
    const replyDiv = document.createElement('div');

    replyItems.classList.add('reply-items');
    rowTitle.classList.add('row');
    rowAuthorDate.classList.add('row')
    author.classList.add('post-author');
    date.classList.add('post-date');
    content.classList.add('post-content');
    rowLikes.classList.add('row', 'post-reactions');
    likesThumb.classList.add('material-symbols-outlined', 'likes', 'likes-tumb');
    likesText.classList.add('post-likes');
    dislikesThumb.classList.add('material-symbols-outlined', 'likes', 'dislikes-tumb');
    dislikesText.classList.add('post-dislikes');
    rowAddRepy.classList.add('row', 'post-addition');
    addReplySymbol.classList.add('material-symbols-outlined', 'likes');
    addReplyText.classList.add('post-addreply');
    repliesInfo.classList.add('post-replies');
    repliesInfo.id = `comment-${comment.id}`

    if (numberOfRepliesForParent !== undefined) {
        if (comment.comment_id === 0) {
            const element = document.getElementById(`post-${comment.post_id}`);
            element.textContent = numberOfRepliesForParent + " replies";
            if (numberOfRepliesForParent > 0 && !element.classList.contains('clickable')) {
                element.classList.add('clickable');
                const parentReplyDiv = parent.querySelector('.replies');
                element.addEventListener("click", () => openReplies(comment.id, "post", parentFormattedID, parentReplyDiv));
            }
        } else if (comment.post_id === 0){
            const element = document.getElementById(`comment-${comment.comment_id}`);
            element.textContent = numberOfRepliesForParent + " replies";
            if (numberOfRepliesForParent > 0 && !element.classList.contains('clickable')) {
                element.classList.add('clickable');
                const parentReplyDiv = parent.querySelector('.replies');
                element.addEventListener("click", () => openReplies(comment.id, "comment", parentFormattedID, parentReplyDiv));
            }
        }
    }

    rowBottom.classList.add('row');
    addReplyDiv.classList.add('add-reply');
    replyDiv.classList.add('replies');

    author.textContent = comment.user.username;
    date.textContent = formatDate(comment.created_at);
    content.textContent = comment.description;
    likesThumb.textContent = "thumb_up";
    likesText.textContent = comment.number_of_likes;
    dislikesThumb.textContent = "thumb_down";
    dislikesText.textContent = comment.number_of_dislikes;
    addReplySymbol.textContent = "chat_bubble"
    addReplyText.textContent = "add reply"
    repliesInfo.textContent = comment.repliesCount + " replies";

    if (comment.liked) likesThumb.style.color = "green"
    if (comment.disliked) dislikesThumb.style.color = "red"

    likesThumb.addEventListener("click", () => handleLike(comment.id, "comment"));
    dislikesThumb.addEventListener("click", () => handleDislike(comment.id, "comment"));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, comment.id, "comment"))

    if (comment.repliesCount > 0) {
        repliesInfo.classList.add('clickable');
        repliesInfo.addEventListener("click", () => openReplies(comment.id, "comment", formattedID, replyDiv));
    }

    rowAuthorDate.appendChild(author);
    rowAuthorDate.appendChild(date);
    replyItems.appendChild(rowAuthorDate);
    replyItems.appendChild(rowTitle);
    replyItems.appendChild(content);
    rowLikes.appendChild(likesThumb);
    rowLikes.appendChild(likesText);
    rowLikes.appendChild(dislikesThumb);
    rowLikes.appendChild(dislikesText);
    rowBottom.appendChild(rowLikes);
    rowAddRepy.appendChild(addReplySymbol);
    rowAddRepy.appendChild(addReplyText);
    rowBottom.appendChild(rowAddRepy);
    rowBottom.appendChild(repliesInfo);
    replyItems.appendChild(rowBottom);
    replyItems.appendChild(addReplyDiv);

    newReply.appendChild(replyItems);
    newReply.appendChild(replyDiv);

    const replyDivs = parent.querySelector(".replies")
    replyDivs.prepend(newReply);
}