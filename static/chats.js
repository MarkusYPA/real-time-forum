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