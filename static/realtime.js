import { fetchPosts, removeLastCategory, sendPost, updateCategory } from "./posts.js";
import { addPostToFeed, addReplyToParent } from "./createposts.js";
import { createUserList, getUsersListing, showChat } from "./chats.js";

export const feed = document.getElementById('posts-feed');
export let ws;

function chatMessages(msg) {
    if (msg.msgType == "listOfChat") createUserList(msg);
    if (msg.msgType == "updateClients") getUsersListing();

    if (msg.msgType == "sendMessage") {
        console.log(msg.message)
    }

    if (msg.msgType == "showMessages") {
        console.log(msg);
        showChat(msg);
    }
}

function forumMessages(msg) {
    let postToModify;
    let replyToModify;

    if (msg.updated && msg.msgType === "post") {
        postToModify = document.getElementById(`postid${msg.post.id}`);
    }
    if (msg.updated && msg.msgType === "comment") {
        replyToModify = document.getElementById(`replyid${msg.comment.id}`);
    }


    // try to find parent for new reply
    let parentForReply
    if (!msg.updated && msg.msgType === "comment") {

        if (msg.comment.post_id === 0) {
            parentForReply = document.getElementById(`replyid${msg.comment.comment_id}`);
        } else {
            parentForReply = document.getElementById(`postid${msg.comment.post_id}`);
        }
    }

    // modify or add
    // Modification means add/remove likes/dislikes
    if (msg.msgType === "post" && postToModify) {
        const likesText = postToModify.querySelector(".post-likes");
        likesText.textContent = msg.post.number_of_likes;
        const dislikesText = postToModify.querySelector(".post-dislikes");
        dislikesText.textContent = msg.post.number_of_dislikes;

        const thumbUp = postToModify.querySelector(".likes-tumb");
        const thumbDown = postToModify.querySelector(".dislikes-tumb");

        changeLikeColor(thumbUp, thumbDown, msg.isLikeAction, msg.post.liked, msg.post.disliked)

        // msg.post.liked ? thumbUp.style.color = "green" : thumbUp.style.color = "var(--text1)";
        // msg.post.disliked ? thumbDown.style.color = "red" : thumbDown.style.color = "var(--text1)";

    } else if (msg.msgType == "comment" && replyToModify) {
        const likesText = replyToModify.querySelector(".post-likes");
        likesText.textContent = msg.comment.number_of_likes;
        const dislikesText = replyToModify.querySelector(".post-dislikes");
        dislikesText.textContent = msg.comment.number_of_dislikes;

        const thumbUp = replyToModify.querySelector(".likes-tumb");
        const thumbDown = replyToModify.querySelector(".dislikes-tumb");
        changeLikeColor(thumbUp, thumbDown, msg.isLikeAction, msg.comment.liked, msg.comment.disliked)
    } else if (parentForReply) {
        // open existing replies, newest on top
        addReplyToParent(parentForReply.id, msg.comment, msg.numberOfReplies);
    } else {
        addPostToFeed(msg.post);
    }
}

// WebSocket message handler
function handleWebSocketMessage(event) {
    const msg = JSON.parse(event.data);

    if (
        msg.msgType == "listOfChat" ||
        msg.msgType == "updateClients" ||
        msg.msgType == "sendMessage" ||
        msg.msgType == "showMessages"
    ) {
        chatMessages(msg)
    }

    if (msg.msgType == "post" || msg.msgType == "comment") {
        forumMessages(msg)
    }
};

function changeLikeColor(thumbUp, thumbDown, isLikeAction, liked, disliked) {
    const computedThumbUpColor = window.getComputedStyle(thumbUp).color;
    const computedThumbDownColor = window.getComputedStyle(thumbDown).color;
    // Check if it's already active and needs to be toggled off
    if (computedThumbUpColor === "rgb(0, 128, 0)" && isLikeAction) { // Green
        thumbUp.style.color = "var(--text1)";
    } else if (computedThumbUpColor !== "rgb(0, 128, 0)" && liked) {
        thumbUp.style.color = "green";
    }

    // Fixing the incorrect element update
    if (computedThumbDownColor === "rgb(255, 0, 0)" && isLikeAction) { // Red
        thumbDown.style.color = "var(--text1)";
    } else if (computedThumbDownColor !== "rgb(255, 0, 0)" && disliked) {
        thumbDown.style.color = "red";
    }
}

function openRegisteration() {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('register-section').style.display = 'flex';
}

function startUp(data) {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('forum-section').style.display = 'block';
    document.getElementById('chat-section').style.display = 'none';
    fetchPosts(0);
    // make server respond with list of clients
    getUsersListing();

    ws = new WebSocket(`ws://localhost:8080/ws?session=${data.token}`);
    ws.onmessage = event => handleWebSocketMessage(event);
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
                startUp(data);
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
            document.getElementById('login-section').style.display = 'flex';
            document.getElementById('forum-section').style.display = 'none';
            document.getElementById('chat-section').style.display = 'none';
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
    fetch('/api/session', { method: 'GET', credentials: 'include' })  // New endpoint to check session
        .then(res => res.json())
        .then(data => {
            if (data.loggedIn) {
                startUp(data);
            }
        });
});