import { handleDislike, handleLike, openAndSendReply, openReplies } from "./posts.js";
import { feed } from "./realtime.js";

// Function to add a post to the page
export function addPostToFeed(post) {
    const newPost = document.createElement('div');
    newPost.className = 'post';
    const formattedID = `postid${post.id}`;
    newPost.id = formattedID;

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
    const categories = document.createElement('span');
    const replies = document.createElement('div')

    rowTitle.classList.add('row', 'spread');
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
    categories.classList.add('post-categories');
    replies.classList.add('replies')

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
    categories.textContent = post.categories.join(', ');

    likesThumb.addEventListener("click", () => handleLike(post.id));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, post.id))
    if (post.repliescount > 0) {
        repliesInfo.classList.add('clickable');
        repliesInfo.addEventListener("click", () => openReplies(post.id, formattedID, replies));
    }

    rowTitle.appendChild(title);
    rowAuthorDate.appendChild(author);
    rowAuthorDate.appendChild(date);
    rowTitle.appendChild(rowAuthorDate);
    newPost.appendChild(rowTitle);
    newPost.appendChild(content);
    rowLikes.appendChild(likesThumb);
    rowLikes.appendChild(likesText);
    rowLikes.appendChild(dislikesThumb);
    rowLikes.appendChild(dislikesText);
    rowBottom.appendChild(rowLikes);
    rowAddRepy.appendChild(addReplySymbol);
    rowAddRepy.appendChild(addReplyText);
    rowBottom.appendChild(rowAddRepy);
    rowBottom.appendChild(repliesInfo);
    rowBottom.appendChild(categories);
    newPost.appendChild(rowBottom);
    newPost.appendChild(replies);

    feed.prepend(newPost);
}

export function addReplyToParent(parentFormattedID, post) {
    const parent = document.getElementById(parentFormattedID);

    const newReply = document.createElement('div');
    newReply.className = 'reply';
    const formattedID = `replyid${post.id}`;
    newReply.id = formattedID;

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
    const replies = document.createElement('div')

    rowTitle.classList.add('row', 'spread');
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
    replies.classList.add('replies')

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

    likesThumb.addEventListener("click", () => handleLike(post.id));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id));
    rowAddRepy.addEventListener("click", () => openAndSendReply(formattedID, post.id))
    if (post.repliescount > 0) {
        repliesInfo.classList.add('clickable');
        repliesInfo.addEventListener("click", () => openReplies(post.id, formattedID, replies));
    }

    rowAuthorDate.appendChild(author);
    rowAuthorDate.appendChild(date);
    rowTitle.appendChild(rowAuthorDate);
    newReply.appendChild(rowTitle);
    newReply.appendChild(content);
    rowLikes.appendChild(likesThumb);
    rowLikes.appendChild(likesText);
    rowLikes.appendChild(dislikesThumb);
    rowLikes.appendChild(dislikesText);
    rowBottom.appendChild(rowLikes);
    rowAddRepy.appendChild(addReplySymbol);
    rowAddRepy.appendChild(addReplyText);
    rowBottom.appendChild(rowAddRepy);
    rowBottom.appendChild(repliesInfo);
    newReply.appendChild(rowBottom);
    newReply.appendChild(replies);

    const replyDivs = parent.querySelector(".replies")
    replyDivs.prepend(newReply);
}