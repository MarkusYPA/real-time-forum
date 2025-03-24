import { handleDislike, handleLike, openAndSendReply, openReplies } from "./posts.js";
import { feed } from "./realtime.js";

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
    //const categories = document.createElement('span');
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
    likesThumb.classList.add('material-symbols-outlined', 'likes');
    likesText.classList.add('post-likes');
    dislikesThumb.classList.add('material-symbols-outlined', 'likes');
    dislikesText.classList.add('post-dislikes');
    rowAddRepy.classList.add('row', 'post-addition');
    addReplySymbol.classList.add('material-symbols-outlined', 'likes');
    addReplyText.classList.add('post-addreply');
    repliesInfo.classList.add('post-replies');
    rowBottom.classList.add('row');
    //categories.classList.add('post-categories');
    addReplyDiv.classList.add('add-reply');
    replyDiv.classList.add('replies');

    title.textContent = post.title;
    author.textContent = post.author;
    date.textContent = post.date;
    content.textContent = post.content;
    likesThumb.textContent = "thumb_up";
    likesText.textContent = post.likes;
    dislikesThumb.textContent = "thumb_down";
    dislikesText.textContent = post.dislikes;
    addReplySymbol.textContent = "chat_bubble"
    addReplyText.textContent = "add reply"
    repliesInfo.textContent = post.repliescount + " replies";
    //categories.textContent = post.categories.join(', ');

    likesThumb.addEventListener("click", () => handleLike(post.id, "post"));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id, "post"));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, post.id, "post"))

    if (post.repliescount > 0) {
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

    //rowBottom.appendChild(categories);
    post.categories.forEach(cat => {
        const category = document.createElement('span');
        category.classList.add('post-categories');
        category.textContent = cat;
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

export function addReplyToParent(parentFormattedID, post) {
    const parent = document.getElementById(parentFormattedID);

    const newReply = document.createElement('div');
    newReply.className = 'reply';
    const formattedID = `replyid${post.id}`;
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
    likesThumb.classList.add('material-symbols-outlined', 'likes');
    likesText.classList.add('post-likes');
    dislikesThumb.classList.add('material-symbols-outlined', 'likes');
    dislikesText.classList.add('post-dislikes');
    rowAddRepy.classList.add('row', 'post-addition');
    addReplySymbol.classList.add('material-symbols-outlined', 'likes');
    addReplyText.classList.add('post-addreply');
    repliesInfo.classList.add('post-replies');
    rowBottom.classList.add('row');
    addReplyDiv.classList.add('add-reply');
    replyDiv.classList.add('replies');

    author.textContent = post.author;
    date.textContent = post.date;
    content.textContent = post.content;
    likesThumb.textContent = "thumb_up";
    likesText.textContent = post.likes;
    dislikesThumb.textContent = "thumb_down";
    dislikesText.textContent = post.dislikes;
    addReplySymbol.textContent = "chat_bubble"
    addReplyText.textContent = "add reply"
    repliesInfo.textContent = post.repliescount + " replies";

    likesThumb.addEventListener("click", () => handleLike(post.id, "comment"));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id, "comment"));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, post.id, "comment"))

    if (post.repliescount > 0) {
        repliesInfo.classList.add('clickable');
        repliesInfo.addEventListener("click", () => openReplies(post.id, "comment", formattedID, replyDiv));
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