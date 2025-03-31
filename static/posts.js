import { addReplyToParent } from "./createposts.js";
import { addPostToFeed } from "./createposts.js";
import { feed, toggleInput, logout } from "./realtime.js";

// Fetch initial posts
export function fetchPosts(categoryId) {
    feed.innerHTML = "";
    fetch(`/api/posts?categoryid=${categoryId}`)
        .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
        .then(data => {
            if (data.success) {
                if (data.posts && Array.isArray(data.posts)) {
                    data.posts.forEach(addPostToFeed);
                }
            } else {
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in.";
                if (data.message && data.message == "Not logged in") {
                    logout();
                }
            }
        });
}

export function openReplies(parentID, parentType, formattedID, repliesDiv) {
    const replies = repliesDiv.querySelectorAll(".reply");

    if (replies.length != 0) {
        replies.forEach(reply => reply.remove())
        return;
    }

    fetch(`/api/replies?parentID=${parentID}&parentType=${parentType}`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (data.success) {
                if (data.comments && Array.isArray(data.comments)) {
                    data.comments.forEach(comment => addReplyToParent(formattedID, comment));
                }
            } else {
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in.";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log("error opening replies")
                }
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
        .then(data => {
            if (!data.success) {
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in.";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log("error liking post")
                }
            }
        })
        .catch(err => console.error("Error liking post:", err));
}

export function handleDislike(postID, postType) {
    fetch(`/api/dislike?postType=${postType}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        .then(data => {
            if (!data.success) {
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log("error disliking post")
                }
            }
        })        
        .catch(err => console.error("Error disliking post:", err));
}

export function openAndSendReply(formattedID, parentID, parentType) {
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

        fetch(`/api/addreply?parentType=${parentType}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content, parentid: parentID })
        })
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    replyContainer.remove();
                } else {
                    document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in";
                    if (data.message && data.message == "Not logged in") {
                        console.log(data.message);
                        logout();
                    } else {
                        console.log("error adding reply")
                    }
                }
            });
    });
}

let categories = []; // selected categories
let categoryIds = [];

// Send a new post to the server
export async function sendPost() {
    const titleInput = document.getElementById('postTitle');
    const contentInput = document.getElementById('postInput');
    const errorMessage = document.getElementById('errorMessageFeed');

    const title = titleInput.value.trim();
    const content = contentInput.value.trim();

    if (!content || !title || categoryIds.length < 1) {
        errorMessage.style.display = 'block';
        errorMessage.textContent = "Title, content and more than 0 categories required"
        return;
    }

    errorMessage.textContent = '';
    errorMessage.style.display = 'none';

    await fetch('/api/posts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, categoryIds })
    })
        .then(res => res.json())
        .then(data => {
            if (!data.success) {
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message);
                    logout();
                } else {
                    console.log("error getting posts")
                }
            }
        });



    // Clear input fields
    titleInput.value = '';
    contentInput.value = '';
    categories = [];
    categoryIds = [];
    document.getElementById('categories').innerHTML = '';
    toggleInput();
}

export function updateCategory() {
    const select = document.getElementById("category-selector");
    const selectedCategoryName = select.value.split("_")[0];
    const selectedCategoryID = Number(select.value.split("_")[1]);

    if (selectedCategoryName && !categories.includes(selectedCategoryName)) {
        categories.push(selectedCategoryName);
        categoryIds.push(selectedCategoryID)
        renderCategories();
    }
    select.selectedIndex = 0; // Reset dropdown selection
}

export function removeLastCategory() {
    if (categories.length > 0) {
        categories.pop(); // Remove the last added category
        categoryIds.pop();
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