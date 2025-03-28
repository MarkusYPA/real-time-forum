export function getUsersListing() {
    fetch(`/api/userslist`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                // deal with error
            }
        });
}

export function sendMessage(UserUUID, ChatUUID, content){
    fetch(`/api/sendmessage?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`,{
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({content})
    })
    .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
    .then(data => {
        if (!data.success) {
            // deal with error
        }
    });
}

export function showMessages(ChatUUID, UserUUID, numberOfMessages){
    console.log("trying to show messages")

    fetch(`/api/showmessages?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`,{
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({numberOfMessages})
    })
    .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
    .then(data => {
        if (!data.success) {
            console.log(data.message);
            // deal with error
        }
    });
}

function fillUser(user, userList){
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
            const userUUID = user.userUuid;
            const chatUUID = "";
            if (user.chatUUID.Valid) chatUUID = user.chatUUID.String;

            showMessages(chatUUID, userUUID, 10)
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

export function showChat(privateMessages) {
    document.getElementById('forum-section').style.display = 'none';
    const chat =document.getElementById('chat-section')
    chat.style.display = 'flex';

    const chatContainer = document.createElement('div');
    chatContainer.classList.add('chat-container');

    chat.appendChild(chatContainer);

}