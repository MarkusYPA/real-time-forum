const feed = document.getElementById('feed');
const ws = new WebSocket('ws://localhost:8080/ws');

// Fetch initial posts
fetch('/api/posts')
    .then(res => res.json())
    .then(posts => {
        if (!Array.isArray(posts)) posts = []; // Ensure it's an array
        posts.forEach(addPostToFeed);
    })
    .catch(error => console.error('Error fetching posts:', error));

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