const feed = document.getElementById('postsFeed');
const ws = new WebSocket('ws://localhost:8080/ws');


function openRegisteration() {
    document.getElementById('loginSection').style.display = 'none';
    document.getElementById('registerSection').style.display = 'block';
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
                document.getElementById('loginSection').style.display = 'none';
                document.getElementById('forumSection').style.display = 'block';
                fetchPosts();
            } else {
                document.getElementById('errorMessageLogin').textContent = data.message || "Login failed!";
            }
        });
}

function openLogin() {
    document.getElementById('registerSection').style.display = 'none';
    document.getElementById('loginSection').style.display = 'block';
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
    fetch('/api/logout', { method: 'POST' })
        .then(() => {
            document.getElementById('loginSection').style.display = 'block';
            document.getElementById('forumSection').style.display = 'none';
        });
}

// Fetch initial posts
function fetchPosts() {
    feed.innerHTML = "";
    fetch('/api/posts')
        .then(res => res.json().then(data => ({ success: res.ok, ...data }))) // Merge res.ok into data
        .then(data => {
            if (data.success) {
                if (data.posts && !Array.isArray(data.posts)) data.posts.forEach(addPostToFeed);
            } else {
                document.getElementById('errorMessageFeed').textContent = data.message || "Error loading posts.";
            }
        });
}

// WebSocket message handler
ws.onmessage = event => {
    const post = JSON.parse(event.data);
    addPostToFeed(post);
};

// Function to add a post to the page
function addPostToFeed(post) {
    const div = document.createElement('div');
    div.className = 'post';
    div.textContent = post.content;
    feed.prepend(div);
}

// Send a new post to the server
function sendPost() {
    const input = document.getElementById('postInput');
    const content = input.value.trim();
    if (!content) return;

    fetch('/api/posts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content })
    })
        .then(res => res.json())
        .then(data => {
            if (!data.success) {
                document.getElementById('loginSection').style.display = 'block';
                document.getElementById('forumSection').style.display = 'none';
            }
        });

    input.value = '';  // Clear input
}

addEventListener("DOMContentLoaded", function () {

    // Show forumSection directly if user has a valid session
    fetch('/api/session', { method: 'GET' })  // New endpoint to check session
        .then(res => res.json())
        .then(data => {
            if (data.loggedIn) {
                document.getElementById('loginSection').style.display = 'none';
                document.getElementById('forumSection').style.display = 'block';
                fetchPosts();
            }
        });
});