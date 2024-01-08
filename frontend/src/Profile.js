import React from 'react';

function Profile() {
    // Placeholder for user data
    const user = {
        // Public User Data
        displayName: "User Name",
        personalEmoji: "😊",
        personalLink1: "http://link1.com",
        personalLink2: "http://link2.com",
        personalLink3: "http://link3.com",
        personalLink4: "http://link4.com",
        description: "User description goes here...",
        accountBalance: 100.00,
        // Private Information
    };

    return (
        <div>
            <h1>Profile Page</h1>
            <p><strong>Display Name:</strong> {user.displayName} <button>Edit</button></p>
            <p><strong>Personal Emoji:</strong> {user.personalEmoji} <button>Edit</button></p>
            <p><strong>Personal Links:</strong></p>
            <ul>
                {[1, 2, 3, 4].map((index) => (
                    <li key={index}>
                        {user[`personalLink${index}`]} <button>Edit</button>
                    </li>
                ))}
            </ul>
            <p><strong>Description:</strong> {user.description}</p>
            <p><strong>Account Balance:</strong> ${user.accountBalance.toFixed(2)}</p>



            <button>Change Password</button>
            <button>Regenerate API Key</button>
        </div>
    );
}

export default Profile;
