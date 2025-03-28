export function getUsersListing(){
        fetch(`/api/userslist`)
        .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
        .then(data => {
            if (!data.success) {
                // deal with error
            }
        });
}

export function createUserList(){
    
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
    fetch(`/api/showmessages?UserUUID=${UserUUID}&ChatUUID=${ChatUUID}`,{
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({numberOfMessages})
    })
    .then(res => res.json().catch(() => ({ success: false, message: "Invalid JSON response" }))) // Prevent JSON parse errors
    .then(data => {
        if (!data.success) {
            // deal with error
        }
    });
}

