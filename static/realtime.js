const feed = document.getElementById('postsFeed');
const ws = new WebSocket('ws://localhost:8080/ws');

// WebSocket message handler
ws.onmessage = event => {
    const post = JSON.parse(event.data);

    // try to find matching post
    postToModify = document.getElementById(post.id)

    // modify or add
    if (postToModify) {
        const likesText = postToModify.querySelector(".post-likes");
        likesText.textContent = post.likes;
        const dislikesText = postToModify.querySelector(".post-dislikes");
        dislikesText.textContent = post.dislikes;
    } else {
        addPostToFeed(post);
    }
};


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
    feed.innerHTML = "";
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
                if (data.posts && Array.isArray(data.posts)) {
                    data.posts.forEach(addPostToFeed);
                }
            } else {
                document.getElementById('errorMessageFeed').textContent = data.message || "Error loading posts.";
            }
        });
}

function handleLike(postID) {
    fetch("/api/like", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Liked post ${postID}:`, data))
        .catch(err => console.error("Error liking post:", err));
}

function handleDislike(postID) {
    fetch("/api/dislike", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ postID })
    })
        .then(res => res.json())
        //.then(data => console.log(`Disliked post ${postID}:`, data))
        .catch(err => console.error("Error disliking post:", err));
}

// Function to add a post to the page
function addPostToFeed(post) {
    const newPost = document.createElement('div');
    newPost.className = 'post';
    newPost.id = post.id

    const rowTitle = document.createElement('div');
    const title = document.createElement('div');
    const rowAuthorDate = document.createElement('div');
    const author = document.createElement('span');
    const date = document.createElement('span');
    const content = document.createElement('div');
    const rowCatsReactions = document.createElement('div');
    const categories = document.createElement('span');
    const rowLikes = document.createElement('div');
    const likesThumb = document.createElement('span');
    const likesText = document.createElement('span');
    const dislikesThumb = document.createElement('span');
    const dislikesText = document.createElement('span');

    rowTitle.classList.add('row', 'spread');
    title.classList.add('post-title');
    rowAuthorDate.classList.add('row')
    author.classList.add('post-author');
    date.classList.add('post-date');
    content.classList.add('post-content');
    rowCatsReactions.classList.add('row', 'spread');
    categories.classList.add('post-categories');
    rowLikes.classList.add('row');
    likesThumb.classList.add('material-symbols-outlined', 'likes');
    likesText.classList.add('post-likes');
    dislikesThumb.classList.add('material-symbols-outlined', 'likes');
    dislikesText.classList.add('post-dislikes');

    title.textContent = post.title;
    author.textContent = post.author;
    date.textContent = post.date;
    content.textContent = post.content;
    categories.textContent = post.categories.join(', ');
    likesThumb.textContent = "thumb_up";
    likesText.textContent = post.likes;
    dislikesThumb.textContent = "thumb_down";
    dislikesText.textContent = post.dislikes;    

    likesThumb.addEventListener("click", () => handleLike(post.id));
    dislikesThumb.addEventListener("click", () => handleDislike(post.id));

    rowTitle.appendChild(title);
    rowAuthorDate.appendChild(author);
    rowAuthorDate.appendChild(date);
    rowTitle.appendChild(rowAuthorDate);
    newPost.appendChild(rowTitle);
    newPost.appendChild(content);
    rowCatsReactions.appendChild(categories);
    rowLikes.appendChild(likesThumb);
    rowLikes.appendChild(likesText);
    rowLikes.appendChild(dislikesThumb);
    rowLikes.appendChild(dislikesText);
    rowCatsReactions.appendChild(rowLikes);
    newPost.appendChild(rowCatsReactions);

    feed.prepend(newPost);
}

// Send a new post to the server
function sendPost() {
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

function updateCategory() {
    const select = document.getElementById("categorySelector");
    const selectedCategory = select.value;

    if (selectedCategory && !categories.includes(selectedCategory)) {
        categories.push(selectedCategory);
        renderCategories();
    }
    select.selectedIndex = 0; // Reset dropdown selection
}

function removeLastCategory() {
    if (categories.length > 0) {
        categories.pop(); // Remove the last added category
        renderCategories();
    }
}

function renderCategories() {
    const categoriesDiv = document.getElementById("categories");
    categoriesDiv.textContent = categories.join(' ');
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