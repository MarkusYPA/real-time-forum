# real-time-forum

## Overview
A forum web application that allows users to register, log in, create posts, comment on posts, and send private messages in real-time. The application is built using:

- **SQLite** for storing user and forum data
- **Golang** for backend processing and WebSocket handling
- **JavaScript** for frontend interactivity and WebSocket communication
- **HTML** for structuring the webpage
- **CSS** for styling the webpage

The forum operates as a **Single Page Application (SPA)**, meaning all page changes are handled dynamically using JavaScript.

## Features

### Registration and Login
- Users must register before accessing the forum.
- Registration requires the following details:
  - **Username**
  - **Age**
  - **Gender**
  - **First Name**
  - **Last Name**
  - **Email**
  - **Password**
- Users can log in using either their **username** or **email** combined with their **password**.
- Users can log out from any page on the forum.

### Posts and Comments
- Users can:
  - Create posts categorized by topics.
  - Comment on existing posts.
  - View posts in a feed display.
  - See comments only after clicking on a post.

### Private Messaging
- Users can send private messages to each other.
- A sidebar displays online and offline users.
- The user list is sorted by the latest message sent. If no messages exist, users are sorted alphabetically.
- Clicking a user in the sidebar opens past messages.
- Private messages:
  - Load the last **10 messages** initially.
  - Allow users to scroll up to load **10 more messages** without excessive API calls (**throttling used**).
  - Include a **timestamp** and **sender's name**.
- Messages are **real-time**, using WebSockets to notify users of new messages instantly.


## Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/forum-project.git
   ```
2. Navigate to the project directory:
   ```sh
   cd forum-project
   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```
4. Run the server:
   ```sh
   go run main.go
   ```
5. Open the application in your browser at `http://localhost:8080`

## Usage
- Register a new user and log in.
- Create posts and interact with comments.
- Send private messages and experience real-time updates.

## Authors
- [Mahdi Kheirkhah](https://github.com/mahdikheirkhah)
- [Markus Amberla](https://github.com/MarkusYPA)
