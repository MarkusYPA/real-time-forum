export function getUsersListing() {
    fetch(`/api/userslist`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                // deal with error
            }
        });
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
            const userRow = document.createElement('div');
            userRow.classList.add('row', 'chat-user');
            userRow.id = user.userUuid; // To display new message notification

            const chatSymbol = document.createElement('span');
            chatSymbol.classList.add('material-symbols-outlined');
            chatSymbol.textContent = "chat";
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
                    const userID = user.userUuid;
                    const chatUUID = "";
                    if (user.chatUUID.Valid) chatUUID = user.chatUUID.String;
                    console.log(`User ID: ${userID}, Chat ID: ${chatUUID}`);
                });
            }

            userList.appendChild(userRow)
        });
    }

    const nonAcquaintances = document.createElement('span');
    nonAcquaintances.classList.add('chat-small-title')
    nonAcquaintances.textContent = "No chat yet";
    userList.appendChild(nonAcquaintances)

    if (msg.unchattedUsers) {
        msg.unchattedUsers.forEach(user => {
            const userRow = document.createElement('div');
            userRow.classList.add('row', 'chat-user');
            userRow.id = user.userUuid; // To display new message notification

            const chatSymbol = document.createElement('span');
            chatSymbol.classList.add('material-symbols-outlined');
            chatSymbol.textContent = "chat";
            userRow.appendChild(chatSymbol);

            const name = document.createElement('span');
            name.classList.add('chat-user-name');
            name.textContent = user.username;
            userRow.appendChild(name);

            if (user.isOnline) {
                userRow.classList.add('clickable');
                const status = document.createElement('span');
                status.classList.add('chat-user-status');
                status.textContent = "online";
                userRow.appendChild(status)

                userRow.addEventListener('click', () => {
                    const userID = user.userUuid;
                    const chatUUID = "";
                    if (user.chatUUID.Valid) chatUUID = user.chatUUID.String;
                    
                    console.log(`User ID: ${userID}, Chat ID: ${chatUUID}`);
                });
            }

            userList.appendChild(userRow)
        });
    }

    messages.appendChild(userList)
}