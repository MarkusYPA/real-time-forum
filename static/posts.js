import { addReplyToParent } from "./createposts.js";
import { addPostToFeed } from "./createposts.js";
import { feed } from "./realtime.js";

// Fetch initial posts
export function fetchPosts() {
    feed.innerHTML = "";
    fetch('/api/posts')
        .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
        .then(data => {
            if (data.success) {
                if (data.posts && Array.isArray(data.posts)) {
                    data.posts.forEach(addPostToFeed);
                }
            } else {
                document.getElementById('errorMessageFeed').textContent = data.message || "Error loading posts.";
            }
        });
}

export function openReplies(postID, formattedID, repliesDiv){
    const replies = repliesDiv.querySelectorAll(".reply");
    if (replies.length != 0) {
        replies.forEach( reply => reply.remove())      
        return;
    }

    fetch(`/api/replies?postID=${postID}`)
    .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
    .then(data => {
        if (data.success) {
            if (data.replies && Array.isArray(data.replies)) {
                data.replies.forEach(reply => addReplyToParent(formattedID, reply));
            }
        } else {
            document.getElementById('errorMessageFeed').textContent = data.message || "Error loading posts.";
        }
    });
}

export function handleLike(postID) {
    fetch("/api/like", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Liked post ${postID}:`, data))
        .catch(err => console.error("Error liking post:", err));
}

export function handleDislike(postID) {
    fetch("/api/dislike", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Disliked post ${postID}:`, data))
        .catch(err => console.error("Error disliking post:", err));
}

export function openAndSendReply(formattedID, parentID) {
    const parent = document.getElementById(formattedID);

    // Check if a reply input already exists
    const oldContainer = parent.querySelector('.reply-container');
    if (oldContainer) {
        oldContainer.remove();
        return;
    }

    const replyContainer = document.createElement('div');
    replyContainer.classList.add('reply-container');

    // Textarea for reply content
    const replyInput = document.createElement('textarea');
    replyInput.placeholder = 'Write a reply...';
    replyInput.classList.add('reply-input');

    // Submit button
    const submitButton = document.createElement('button');
    submitButton.textContent = 'Reply';
    submitButton.classList.add('reply-button');

    // Append input and button to container
    replyContainer.appendChild(replyInput);
    replyContainer.appendChild(submitButton);

    const addReplyDiv = parent.querySelector(".add-reply")
    addReplyDiv.appendChild(replyContainer);

    // Handle submit action
    submitButton.addEventListener('click', function () {
        const content = replyInput.value.trim();
        if (!content) return; // Prevent empty replies

        fetch('/api/addreply', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            //body: JSON.stringify({ content, parentid: parentID })
            body: JSON.stringify({ content, parentid: parentID })
        })
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    replyContainer.remove();
                } else {
                    // Write error message to parent error field?
                }
            });
    });
}


// Send a new post to the server
export function sendPost() {
    const titleInput = document.getElementById('postTitle');
    const contentInput = document.getElementById('postInput');
    const categoriesInput = document.getElementById('categories');

    const title = titleInput.value.trim();
    const content = contentInput.value.trim();
    const categories = categoriesInput.textContent.trim().split(/\s+/);

    if (!content || !title) return;

    fetch('/api/posts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, categories })
    })
        .then(res => res.json())
        .then(data => {
            if (!data.success) {
                document.getElementById('loginSection').style.display = 'block';
                document.getElementById('forumSection').style.display = 'none';
            }
        });

    // Clear input fields
    titleInput.value = '';
    contentInput.value = '';
    categoriesInput.innerHTML = '';
}

let categories = []; // selected categories

export function updateCategory() {
    const select = document.getElementById("categorySelector");
    const selectedCategory = select.value;

    if (selectedCategory && !categories.includes(selectedCategory)) {
        categories.push(selectedCategory);
        renderCategories();
    }
    select.selectedIndex = 0; // Reset dropdown selection
}

export function removeLastCategory() {
    if (categories.length > 0) {
        categories.pop(); // Remove the last added category
        renderCategories();
    }
}

function renderCategories() {
    const categoriesDiv = document.getElementById("categories");
    categoriesDiv.textContent = categories.join(' ');
}