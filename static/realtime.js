import { fetchPosts, removeLastCategory, sendPost, updateCategory } from "./posts.js";
import { addPostToFeed, addReplyToParent } from "./createposts.js";
import { addMessageToChat, createUserList, getUsersListing, showChat } from "./chats.js";

export const feed = document.getElementById('posts-feed');
export let ws;

function chatMessages(msg) {
    if (msg.msgType == "listOfChat") createUserList(msg);
    if (msg.msgType == "updateClients") getUsersListing();

    if (msg.msgType == "sendMessage") {
        getUsersListing();
        console.log(msg)
        console.log("Chat message received:", msg.privateMessage.message.content)
        addMessageToChat(msg);
        const chatUUID = document.getElementById(msg.privateMessage.message.chat_uuid)
        if (msg.notification && !chatUUID) {
            showNotification(msg.privateMessage.message.sender_username)
        }
    }

    if (msg.msgType == "showMessages") {
        showChat(msg);
    }
}

function showNotification(sender) {
    let notificationBox = document.getElementById("notificationBox");
    notificationBox.innerHTML = `📩 New message from <b>${sender}</b>`;
    notificationBox.classList.add("show");

    // Hide after 5 seconds
    setTimeout(() => {
        notificationBox.classList.remove("show");
    }, 5000);
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

    // Add or modify (add/remove likes/dislikes)
    if (msg.msgType === "post" && postToModify) {
        const likesText = postToModify.querySelector(".post-likes");
        likesText.textContent = msg.post.number_of_likes;
        const dislikesText = postToModify.querySelector(".post-dislikes");
        dislikesText.textContent = msg.post.number_of_dislikes;

        const thumbUp = postToModify.querySelector(".likes-tumb");
        const thumbDown = postToModify.querySelector(".dislikes-tumb");
        changeLikeColor(thumbUp, thumbDown, msg.isLikeAction, msg.post.liked, msg.post.disliked)

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
    document.getElementById('forum-container').style.display = 'block';

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
                document.getElementById('errorMessageLogin').textContent = data.message || "Not logged in";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message);
                    logout();
                } else {
                    console.log("error logging in")
                }
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
                document.getElementById('errorMessageLogin').textContent = "User registered succesfully!";                         
            } else {
                document.getElementById('errorMessageRegister').textContent = "Registration failed!";
            }
        });
}

export function logout() {
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

function showForum() {
    document.getElementById('forum-container').style.display = 'block';
    document.getElementById('chat-section').style.display = 'none';
    document.getElementById('profile-section').style.display = 'none';
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
        newCat.addEventListener('click', () => {
            const catButtons = document.getElementsByClassName('view-category');
            Array.from(catButtons).forEach((cb)=>cb.classList.remove('highlight'));
            newCat.classList.add('highlight');
            showCategory(categoryIds[i]);
        });
        if (i==0) {
            newCat.classList.add('highlight');
        }

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

    await fetch('/api/category', { method: 'GET' })
        .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
        .then(data => {
            if (data.success) {
                if (data.categories && Array.isArray(data.categories)) {
                    data.categories.forEach(cat => addCategoryToSelector(cat));
                }
            } else {
                document.getElementById('errorMessageFeed').textContent = data.message || "Error loading categories.";
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message);
                    logout();
                } else {
                    console.log("error loading categories")
                }
            }
        });

    populateCategoryViews(categoryNames, categoryIds);
}

async function myProfile(){
    await fetch('/api/myprofile', { method: 'GET' })
    .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
    .then(data => {
        if (data.success) {
            if (data.user) {
                showUserProfile(data.user);
            }
        } else {
            document.getElementById('errorMessageFeed').textContent = data.message || "Error viewing profile.";
            if (data.message && data.message == "Not logged in") {
                console.log(data.message);
                logout();
            } else {
                console.log("error viewing profile")
            }
        }
    });
}

function showUserProfile(user){
      document.getElementById('forum-container').style.display = 'none';
      document.getElementById('chat-section').style.display = 'none';

        const profile = document.getElementById('profile-section')
        profile.style.display = 'flex';
    
        let profileContainer = document.querySelector('.profile-container');
        if (!profileContainer) {
            profileContainer = document.createElement('div');
            profileContainer.classList.add('profile-container');
        } else {
            profileContainer.innerHTML = '';
        }
        profileContainer.id = '';
        // append early so chatContainer can be found in createChatBubble()
    
        const profileTitle = document.createElement('div');
        profileTitle.classList.add('profile-title');
        profileTitle.textContent = user.username + ' profile: ';
    
        const information = document.createElement('div');
        information.classList.add('information');
        information.id = user.uuid; // id to find correct chat
        information.innerHTML = `
        <p>first name: ${user.firstName}</p> <p>last name: ${user.lastName}</p> <p>Age: ${user.age}</p> <p>gender: ${user.gender}</p> <p>email: ${user.email}</p>`
    
        profileContainer.appendChild(profileTitle);
        profileContainer.appendChild(information);
        profile.appendChild(profileContainer);
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
    document.querySelector('#page-title').addEventListener('click', showForum);
    document.querySelector('#my-profile-button').addEventListener('click', myProfile);

    fetchCategories();

    // Show forum-section directly if user has a valid session
    fetch('/api/session', { method: 'GET', credentials: 'include' })
        .then(res => res.json())
        .then(data => {
            if (data.loggedIn) {
                startUp(data);
            }
        });
});
