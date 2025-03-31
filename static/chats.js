import { formatDate } from "./createposts.js";
import { logout } from "./realtime.js";

let messagesAmount = 10;
let previousScrollPosition = 0;

export function getUsersListing() {
    fetch(`/api/userslist`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log('error getting user list')
                }
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
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log('error processing message')
                }
            } else {
                console.log(data.message)
            }
        });
}

export function showMessages(ChatUUID, UserUUID, numberOfMessages) {
    fetch(`/api/showmessages?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ numberOfMessages })
    })
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                if (data.message && data.message == "Not logged in") {
                    console.log(data.message)
                    logout();
                } else {
                    console.log('error showing messages')
                }
            }
        });
}


function fillUser(user, userList, hasChat) {
    const userRow = document.createElement('div');
    userRow.classList.add('row', 'chat-user');
    userRow.id = 'listedUser' + user.userUuid; // To find for new message notification

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

    userRow.setAttribute("Age", user.user.age);
    userRow.setAttribute("Lastname", user.user.lastName);
    userRow.setAttribute("Firstname", user.user.firstName);
    userRow.setAttribute("Gender", user.user.gender);
    userRow.setAttribute("LastTimeOnline", formatDate(user.user.lastTimeOnline));

    if (user.isOnline || hasChat) {
        userRow.classList.add('clickable');

        if (user.isOnline) {
            userRow.setAttribute("LastTimeOnline", 'Now')
            const status = document.createElement('span');
            status.classList.add('chat-user-status');
            status.textContent = "online";
            userRow.appendChild(status)
        }

        userRow.addEventListener('click', () => {
            let chatUUID = "";
            if (user.chatUUID.Valid) chatUUID = user.chatUUID.String;
            messagesAmount = 10;
            showMessages(chatUUID, user.userUuid, messagesAmount)
        });
    }
    const tooltip = document.getElementById("userTooltip");
    userRow.addEventListener("mouseover", (event) => {
        const age = userRow.getAttribute("age");
        const lastName = userRow.getAttribute("Lastname");
        const firstName = userRow.getAttribute("Firstname");
        const gender = userRow.getAttribute("Gender");
        const lastTimeOnline = userRow.getAttribute("LastTimeOnline");

        tooltip.innerHTML = `first name: <br>${firstName}</br>last name: <br>${lastName}</br>gender: <br>${gender}</br>first name: <br>${firstName}</br><br>age: ${age}<br>last time online: ${lastTimeOnline}`;
        tooltip.style.display = "block";
        tooltip.style.left = event.pageX + 10 + "px";
        tooltip.style.top = event.pageY + 10 + "px";
    });

    userRow.addEventListener("mousemove", (event) => {
        tooltip.style.left = event.pageX + 10 + "px";
        tooltip.style.top = event.pageY + 10 + "px";
    });

    userRow.addEventListener("mouseleave", () => {
        tooltip.style.display = "none";
    });

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
            fillUser(user, userList, true)
        });
    }

    const nonAcquaintances = document.createElement('span');
    nonAcquaintances.classList.add('chat-small-title')
    nonAcquaintances.textContent = "No chat yet";
    userList.appendChild(nonAcquaintances)

    if (msg.unchattedUsers) {
        msg.unchattedUsers.forEach(user => {
            fillUser(user, userList, false)
        });
    }

    messages.appendChild(userList)
}

function createChatBubble(m, chatMessages, append) {
    const chatContainer = document.querySelector('.chat-container');
    chatContainer.id = m.message.chat_uuid;

    const chatBubble = document.createElement('div');
    chatBubble.classList.add('chat-bubble');
    const messageSender = document.createElement('div');
    messageSender.textContent = m.message.sender_username;
    messageSender.classList.add('chat-bubble-sender');
    const messageContent = document.createElement('div');
    messageContent.textContent = m.message.content;
    const timeAndDate = document.createElement('span');
    timeAndDate.classList.add('chat-bubble-time');
    timeAndDate.textContent = formatDate(m.message.created_at);

    chatBubble.appendChild(messageSender);
    chatBubble.appendChild(messageContent);
    chatBubble.appendChild(timeAndDate);

    if (m.isCreatedBy) {
        chatBubble.classList.add('own-message');
    }
    if (!append) {
        chatMessages.prepend(chatBubble);
    } else {
        chatMessages.appendChild(chatBubble);
    }
}

export function showChat(msg) {
    document.getElementById('forum-container').style.display = 'none';
    document.getElementById('profile-section').style.display = 'none';
    const chat = document.getElementById('chat-section')
    chat.style.display = 'flex';

    let chatContainer = document.querySelector('.chat-container');
    if (!chatContainer) {
        chatContainer = document.createElement('div');
        chatContainer.classList.add('chat-container');
    } else {
        chatContainer.innerHTML = '';
    }
    chatContainer.id = '';
    // append early so chatContainer can be found in createChatBubble()
    chat.appendChild(chatContainer);

    const chatTitle = document.createElement('div');
    chatTitle.classList.add('chat-title');
    chatTitle.textContent = 'Chat with ' + msg.receiverUserName;

    let chatUuid = "";
    const chatMessages = document.createElement('div');
    chatMessages.classList.add('chat-bubbles');
    chatMessages.id = msg.reciverUserUUID; // id to find correct chat
    if (msg.privateMessages && Array.isArray(msg.privateMessages)) {
        chatUuid = msg.privateMessages[0].message.chat_uuid;
        msg.privateMessages.forEach((m) => createChatBubble(m, chatMessages, true))
    }
    chatTitle.classList.add('chat-messages');

    // Add throttled loading of more messages if there are more
    if (msg.allMessagesGot) {
        let isThrottled = false;
        chatMessages.addEventListener('scroll', event => {

            if (isThrottled) return;

            isThrottled = true;
            setTimeout(() => {
                if (chatMessages.scrollTop * -1 >= chatMessages.scrollHeight - chatMessages.clientHeight - 1) {
                    if (chatUuid != '') chatUuid = chatContainer.id;
                    if (chatUuid != '') {
                        messagesAmount += 10;
                        previousScrollPosition = chatMessages.scrollTop;
                        showMessages(chatUuid, msg.reciverUserUUID, messagesAmount)
                    }
                }
                isThrottled = false;
            }, 1000); // Throttle delay
        });
    }



    const chatInput = document.createElement('div');
    chatInput.classList.add('chat-input');
    const chatTextInput = document.createElement('textarea');
    chatTextInput.classList.add('chat-textarea');
    chatTextInput.rows = '3';
    chatInput.appendChild(chatTextInput);
    const chatSendButton = document.createElement('button');
    chatSendButton.textContent = "Send";
    chatSendButton.addEventListener('click', () => {
        const receiverUUID = msg.reciverUserUUID;
        const messageText = chatTextInput.value.trim();
        if (messageText != '') {
            sendMessage(receiverUUID, chatUuid, chatTextInput.value.trim());
            chatTextInput.value = '';
        }
    });
    chatInput.appendChild(chatSendButton);

    chatContainer.appendChild(chatTitle);
    chatContainer.appendChild(chatMessages);
    chatContainer.appendChild(chatInput);

    chatMessages.scrollTop = previousScrollPosition;
}

export function addMessageToChat(msg) {
    let chatMessages = document.getElementById(msg.reciverUserUUID);
    if (!chatMessages) chatMessages = document.getElementById(msg.uuid);
    if (chatMessages) createChatBubble(msg.privateMessage, chatMessages, false)
}