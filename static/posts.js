import { addReplyToParent } from "./createposts.js";
import { addPostToFeed } from "./createposts.js";
import { feed, toggleInput } from "./realtime.js";

// Fetch initial posts
export function fetchPosts(categoryId) {
    feed.innerHTML = "";

    //fetch('/api/posts')
    fetch(`/api/posts?categoryid=${categoryId}`)
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

export function openReplies(parentID, postType, formattedID, repliesDiv){
    const replies = repliesDiv.querySelectorAll(".reply");
    if (replies.length != 0) {
        replies.forEach( reply => reply.remove())      
        return;
    }

    fetch(`/api/replies?parentID=${parentID}&postType=${postType}`)
    //.then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
    .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
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

export function handleLike(postID, postType) {
    fetch(`/api/like?postType=${postType}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Liked post ${postID}:`, data))
        .catch(err => console.error("Error liking post:", err));
}

export function handleDislike(postID, postType) {
    fetch(`/api/dislike?postType=${postType}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Disliked post ${postID}:`, data))
        .catch(err => console.error("Error disliking post:", err));
}

export function openAndSendReply(formattedID, parentID, postType) {
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
    replyInput.rows = 6;
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

        fetch(`/api/addreply?postType=${postType}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
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

let categories = []; // selected categories
let categoriIds = [];

// Send a new post to the server
export function sendPost() {
    const titleInput = document.getElementById('postTitle');
    const contentInput = document.getElementById('postInput');

    const title = titleInput.value.trim();
    const content = contentInput.value.trim();

    if (!content || !title) return;

    fetch('/api/posts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, categoriIds })
    })
        .then(res => res.json())
        .then(data => {
            if (!data.success) {
                document.getElementById('login-section').style.display = 'block';
                document.getElementById('forum-section').style.display = 'none';
            }
        });

    // Clear input fields
    titleInput.value = '';
    contentInput.value = '';
    categories = [];
    categoriIds = [];
    document.getElementById('categories').innerHTML = '';
    toggleInput();
}

export function updateCategory() {
    const select = document.getElementById("category-selector");
    const selectedCategoryName = select.value.split("_")[0];
    const selectedCategoryID = select.value.split("_")[1];

    if (selectedCategoryName && !categories.includes(selectedCategoryName)) {
        categories.push(selectedCategoryName);
        categoriIds.push(selectedCategoryID)
        renderCategories();
    }
    select.selectedIndex = 0; // Reset dropdown selection
}

export function removeLastCategory() {
    if (categories.length > 0) {
        categories.pop(); // Remove the last added category
        categoriIds.pop();
        renderCategories();
    }
}

function renderCategories() {
    const categoriesDiv = document.getElementById("categories");
    categoriesDiv.innerHTML = "";

    categories.forEach(cat => {
        const category = document.createElement('span');
        category.classList.add('post-categories', 'writing');
        category.textContent = cat;
        categoriesDiv.appendChild(category);
    });
}