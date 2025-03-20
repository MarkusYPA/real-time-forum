const feed = document.getElementById('postsFeed');
const ws = new WebSocket('ws://localhost:8080/ws');

function openRegisteration() {
    document.getElementById('loginSection').style.display = 'none';
    document.getElementById('registerSection').style.display = 'block';
}

function login() {
    const username = document.getElementById('username-login').value.trim();
    const password = document.getElementById('password-login').value.trim();
    fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    })
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                document.getElementById('loginSection').style.display = 'none';
                document.getElementById('forumSection').style.display = 'block';
                fetchPosts();
            } else {
                document.getElementById('errorMessage').textContent = "Login failed!";
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
                //login(); // Auto-login after registering
                openLogin();
            } else {
                document.getElementById('errorMessage').textContent = "Registration failed!";
            }
        });
}

function logout() {
    fetch('/api/logout', { method: 'POST' })
        .then(() => {
            document.getElementById('authSection').style.display = 'block';
            document.getElementById('forumSection').style.display = 'none';
        });
}

// Fetch initial posts
function fetchPosts() {
    fetch('/api/posts')
        .then(res => res.json())
        .then(posts => {
            //const feed = document.getElementById('postsFeed');
            feed.innerHTML = "";
            if (!Array.isArray(posts)) posts = []; // Ensure it's an array
            posts.forEach(addPostToFeed);

            /*             posts.forEach(post => {
                            const div = document.createElement('div');
                            div.textContent = post.content;
                            feed.appendChild(div);
                        }); */
        });
}

// Fetch initial posts
/* fetch('/api/posts')
    .then(res => res.json())
    .then(posts => {
        if (!Array.isArray(posts)) posts = []; // Ensure it's an array
        posts.forEach(addPostToFeed);
    })
    .catch(error => console.error('Error fetching posts:', error)); */

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
    });

    input.value = '';  // Clear input
}