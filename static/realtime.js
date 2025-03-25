import { fetchPosts, removeLastCategory, sendPost, updateCategory } from "./posts.js";
import { addPostToFeed, addReplyToParent } from "./createposts.js";

export const feed = document.getElementById('posts-feed');
export const ws = new WebSocket('ws://localhost:8080/ws');
// WebSocket message handler
ws.onmessage = event => {
    const post = JSON.parse(event.data);


    let postToModify
    let replyToModify

    // try to find matching post or reply
    if (post.posttype == "post") {
        postToModify = document.getElementById(`postid${post.id}`);
    }
    if (post.posttype == "comment") {
        replyToModify = document.getElementById(`replyid${post.id}`);
    }

    console.log(post)

    // try to find parent for new reply
    let parentForReply
    if (!replyToModify && post.parentid && post.posttype == "comment") {
        parentForReply = document.getElementById(`postid${post.parentid}`);
        if (!parentForReply) parentForReply = document.getElementById(`replyid${post.parentid}`);
    }

    // modify or add
    // Modification means add/remove likes/dislikes
    if (post.posttype == "post" && postToModify) {
        const likesText = postToModify.querySelector(".post-likes");
        likesText.textContent = post.likes;
        const dislikesText = postToModify.querySelector(".post-dislikes");
        dislikesText.textContent = post.dislikes;
    } else if (post.posttype == "comment" && replyToModify) {
        const likesText = replyToModify.querySelector(".post-likes");
        likesText.textContent = post.likes;
        const dislikesText = replyToModify.querySelector(".post-dislikes");
        dislikesText.textContent = post.dislikes;
    } else if (parentForReply) {
        // open existing replies, newest on top
        addReplyToParent(parentForReply.id, post);
    } else {
        addPostToFeed(post);
    }
};


function openRegisteration() {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('register-section').style.display = 'flex';
}

function login() {
    const usernameOrEmail = document.getElementById('username-or-email').value.trim();
    const password = document.getElementById('password-login').value.trim();
    fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ usernameOrEmail, password })
    })
        .then(res => res.json()
            .then(data => ({ success: res.ok, ...data })))  // Merge res.ok with data
        .then(data => {
            if (data.success) {
                document.getElementById('login-section').style.display = 'none';
                document.getElementById('forum-section').style.display = 'block';
                fetchPosts();
            } else {
                document.getElementById('errorMessageLogin').textContent = data.message || "Login failed!";
            }
        });
}

function openLogin() {
    document.getElementById('register-section').style.display = 'none';
    document.getElementById('login-section').style.display = 'flex';
}

function registerUser() {
    const username = document.getElementById('username-register').value.trim();
    const age = document.getElementById('age').value;
    const gender = document.getElementById('gender').value;
    const firstName = document.getElementById('firstname').value.trim();
    const lastName = document.getElementById('lastname').value.trim();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password-register').value.trim();

    fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, age, gender, firstName, lastName, email, password })
    })
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                openLogin();
            } else {
                document.getElementById('errorMessageRegister').textContent = "Registration failed!";
            }
        });
}

function logout() {
    feed.innerHTML = "";
    fetch('/api/logout', { method: 'POST' })
        .then(() => {
            document.getElementById('login-section').style.display = 'block';
            document.getElementById('forum-section').style.display = 'none';
        });
}

export function toggleInput() {
    const inputsContainer = document.querySelector('#input-container');
    const inputs = document.querySelector('#hideable-input');
    if (inputs.style.display == "none") {
        inputsContainer.style.backgroundColor = "var(--bg6)";
        inputs.style.display = "flex";

    } else {
        inputsContainer.style.backgroundColor = "";
        inputs.style.display = "none";
    }
}

function populateCategoryViews() {
    const catsDiv = document.querySelector('#view-categories')
    const viewCats = ["all", "aesthetics", "epistemology", "ethics", "logic", "metaphysics", "other"]

    function showCategory(category) {
        fetchPosts(category);
    }

    viewCats.forEach(cat => {
        const newCat = document.createElement('div');
        newCat.classList.add("view-category");
        newCat.textContent = cat;
        newCat.addEventListener('click', () => showCategory(cat));
        catsDiv.appendChild(newCat);
    })
}


addEventListener("DOMContentLoaded", function () {
    document.querySelector('#login-button').addEventListener('click', login);
    document.querySelector('#open-registeration-button').addEventListener('click', openRegisteration);
    document.querySelector('#register-button').addEventListener('click', registerUser);
    document.querySelector('#open-login-button').addEventListener('click', openLogin);
    document.querySelector('#category-selector').addEventListener('change', updateCategory);
    document.querySelector('#remove-category-button').addEventListener('click', removeLastCategory);
    document.querySelector('#send-post-button').addEventListener('click', sendPost);
    document.querySelector('#logout-button').addEventListener('click', logout);
    document.querySelector('#create-post-text').addEventListener('click', toggleInput);
    populateCategoryViews();

    // Show forum-section directly if user has a valid session
    fetch('/api/session', { method: 'GET' })  // New endpoint to check session
        .then(res => res.json())
        .then(data => {
            if (data.loggedIn) {
                document.getElementById('login-section').style.display = 'none';
                document.getElementById('forum-section').style.display = 'block';
                fetchPosts();
            }
        });
});