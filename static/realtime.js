import { fetchPosts, removeLastCategory, sendPost, updateCategory } from "./posts.js";
import { addPostToFeed, addReplyToParent } from "./createposts.js";

export const feed = document.getElementById('posts-feed');
export let ws;
// WebSocket message handler
function handleWebSocketMessage(event) {
    const msg = JSON.parse(event.data);
    console.log("WebSocket message:", msg);

    let postToModify, replyToModify, parentForReply;
    
    if (msg.updated && msg.msgType === "post") {
        postToModify = document.getElementById(`postid${msg.post.id}`);
    }
    if (msg.updated && msg.msgType === "comment") {
        replyToModify = document.getElementById(`replyid${msg.comment.id}`);
    }

    if (!msg.updated && msg.msgType === "comment") {
        if (msg.comment.post_id === 0){
            parentForReply = document.getElementById(`replyid${msg.comment.comment_id}`);
        } else {
            parentForReply = document.getElementById(`postid${msg.comment.post_id}`);
        }
    }

    if (msg.msgType === "post" && postToModify) {
        updatePostLikesDislikes(postToModify, msg.post);
    } else if (msg.msgType === "comment" && replyToModify) {
        updateCommentLikesDislikes(replyToModify, msg.comment);
    } else if (parentForReply) {
        addReplyToParent(parentForReply.id, msg.comment);
    } else {
        addPostToFeed(msg.post);
    }
}

function openRegisteration() {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('register-section').style.display = 'flex';
}
 
async function login() {
    const usernameOrEmail = document.getElementById('username-or-email').value.trim();
    const password = document.getElementById('password-login').value.trim();

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ usernameOrEmail, password }),
            credentials: 'include'
        });

        const data = await response.json();

        if (data.success) {
            console.log("Logged in successfully");
            // Wait until cookie is definitely available
           
            
            // Now safe to connect WebSocket
            ws = new WebSocket('ws://localhost:8080/ws');
            ws.onmessage = event => handleWebSocketMessage(event);
            // ... rest of your WebSocket setup
            fetchPosts(0)
            
        } else {
            document.getElementById('errorMessageLogin').textContent = data.message || "Login failed!";
        }
    } catch (err) {
        console.error("Login error:", err);
    }
}

function waitForCookie(name, intervalMs, maxAttempts) {
    return new Promise((resolve, reject) => {
        let attempts = 0;
        const checkCookie = () => {
            attempts++;
            if (document.cookie.includes(`${name}`)) {
                resolve();
            } else if (attempts >= maxAttempts) {
                reject(new Error(`Cookie ${name} not found after ${maxAttempts} attempts`));
            } else {
                setTimeout(checkCookie, intervalMs);
            }
        };
        checkCookie();
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

function populateCategoryViews(categoryNames, categoryIds) {
    const catsDiv = document.querySelector('#view-categories')

    function showCategory(categoryId) {
        fetchPosts(categoryId);
    }

    for (let i = 0; i < categoryNames.length; i++) {
        const newCat = document.createElement('div');
        newCat.classList.add("view-category");
        newCat.textContent = categoryNames[i];
        newCat.addEventListener('click', () => showCategory(categoryIds[i]));
        catsDiv.appendChild(newCat);
    }
}

async function fetchCategories() {
    const catSelector = document.getElementById("category-selector");
    const categoryNames = ["All"];
    const categoryIds = [0];

    function addCategoryToSelector(category) {
        const opt = document.createElement("option");
        opt.value = category.name + "_" + category.id;
        opt.textContent = category.name;
        catSelector.appendChild(opt);
        categoryNames.push(category.name);
        categoryIds.push(category.id);
    }

    await fetch('/api/category', { method: 'GET' })  // New endpoint to check session
        .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
        .then(data => {
            if (data.success) {
                if (data.categories && Array.isArray(data.categories)) {
                    data.categories.forEach(cat => addCategoryToSelector(cat));
                }
            } else {
                document.getElementById('errorMessageFeed').textContent = data.message || "Error loading categories.";
            }
        });

    populateCategoryViews(categoryNames, categoryIds);
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
    fetchCategories();
    //populateCategoryViews();


    // // Show forum-section directly if user has a valid session
    fetch('/api/session', { method: 'GET', credentials: 'include'})  // New endpoint to check session
        .then(res => res.json())
        .then(data => {
            if (data.loggedIn) {
                document.getElementById('login-section').style.display = 'none';
                document.getElementById('forum-section').style.display = 'block';
                fetchPosts(0);
            }
        });
});