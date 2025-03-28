import { formatDate } from "./createposts.js";

export function getUsersListing() {
    fetch(`/api/userslist`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                // deal with error
            }
        });
}

export function sendMessage(UserUUID, ChatUUID, content) {
    fetch(`/api/sendmessage?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content })
    })
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                // deal with error
                console.log('there is a problem')
            } else {
                console.log(data.message)
            }
        });
}

export function showMessages(ChatUUID, UserUUID, numberOfMessages) {
    console.log("trying to show messages")

    fetch(`/api/showmessages?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ numberOfMessages })
    })
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                console.log(data.message);
                // deal with error
            }
        });
}


function fillUser(user, userList) {
    const userRow = document.createElement('div');
    userRow.classList.add('row', 'chat-user');
    userRow.id = user.userUuid; // To find for new message notification

    // make this visible at new message
    const chatSymbol = document.createElement('span');
    chatSymbol.classList.add('material-symbols-outlined', 'likes');
    chatSymbol.textContent = "chat";
    chatSymbol.style.visibility = "hidden";
    userRow.appendChild(chatSymbol);

    const name = document.createElement('span');
    name.classList.add('chat-user-name');
    name.textContent = user.username;
    userRow.appendChild(name)

    if (user.isOnline) {
        userRow.classList.add('clickable');
        const status = document.createElement('span');
        status.classList.add('chat-user-status');
        status.textContent = "online";
        userRow.appendChild(status)

        userRow.addEventListener('click', () => {
            const chatUUID = "";
            if (user.chatUUID.Valid) chatUUID = user.chatUUID.String;
            showMessages(chatUUID, user.userUuid, 10)
            //console.log(`User ID: ${userUUID}, Chat ID: ${chatUUID}`);
        });
    }

    userList.appendChild(userRow)
}

export function createUserList(msg) {

    const messages = document.getElementById('messaging-container');

    let userList = document.getElementById('user-list');
    if (!userList) {
        userList = document.createElement('div');
        userList.id = 'user-list';
        userList.classList.add('user-list');
    } else {
        userList.innerHTML = '';
    }

    const acquaintances = document.createElement('span');
    acquaintances.classList.add('chat-small-title')
    acquaintances.textContent = "Existing chats";
    userList.appendChild(acquaintances)

    if (msg.chattedUsers) {
        msg.chattedUsers.forEach(user => {
            fillUser(user, userList)
        });
    }

    const nonAcquaintances = document.createElement('span');
    nonAcquaintances.classList.add('chat-small-title')
    nonAcquaintances.textContent = "No chat yet";
    userList.appendChild(nonAcquaintances)

    if (msg.unchattedUsers) {
        msg.unchattedUsers.forEach(user => {
            fillUser(user, userList)
        });
    }

    messages.appendChild(userList)
}

export function showChat(msg) {
    document.getElementById('forum-container').style.display = 'none';
    const chat = document.getElementById('chat-section')
    chat.style.display = 'flex';

    let chatContainer = document.getElementById('chat-container');
    if (!chatContainer) {
        chatContainer = document.createElement('div');
        chatContainer.id = 'chat-container';
        chatContainer.classList.add('chat-container');
    } else {
        chatContainer.innerHTML = '';
    }

    const chatTitle = document.createElement('div');
    chatTitle.classList.add('chat-title');
    chatTitle.textContent = msg.receiverUserName;


    let chatUuid = "";
    const chatMessages = document.createElement('div');
    if (msg.messages && Array.isArray(msg.messages)) {
        chatUuid = msg.messages[0].chat_uuid;

        msg.messages.forEach((m) => {
            const chatBubble = document.createElement('div');
            chatBubble.classList.add('chat-bubble');
            const timeAndDate = document.createElement('span');
            timeAndDate.textContent = formatDate(m.message.created_at);
            const chatContent = document.createElement('div');
            chatContent.textContent = m.message.content;

            chatBubble.appendChild(timeAndDate);
            chatBubble.appendChild(chatContent);

            if (m.isCreatedBy) {
                chatBubble.classList.add('own-message');
            }
            chatMessages.appendChild(chatBubble);
        })
    }
    chatTitle.classList.add('chat-messages');


    const chatInput = document.createElement('div');
    chatTitle.classList.add('chat-input');
    const chatTextInput = document.createElement('textarea');
    chatInput.appendChild(chatTextInput);
    const chatSendButton = document.createElement('button');
    chatSendButton.textContent = "Send";
    chatSendButton.addEventListener('click', () => {
        const receiverUUID = msg.reciverUserUUID;
        sendMessage(receiverUUID, chatUuid, chatTextInput.value.trim());
    });


    chatInput.appendChild(chatSendButton);

    chatContainer.appendChild(chatTitle);
    chatContainer.appendChild(chatMessages);
    chatContainer.appendChild(chatInput);

    chat.appendChild(chatContainer);

}