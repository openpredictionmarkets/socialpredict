import React, { useState, useEffect } from 'react';
import './Profile.css'; // Import the CSS file
import { API_URL } from './config';

function Profile() {
    const [userData, setUserData] = useState(null);

    useEffect(() => {
        // Retrieve the JWT token from localStorage
        const token = localStorage.getItem('token');
        if (!token) {
            console.error("Authentication token not found. Please log in again.");
            return;
        }

        fetch(`${API_URL}/api/v0/user/privateprofile`, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
        })
        .then(response => response.json())
        .then(data => setUserData(data))
        .catch(error => console.error('Error fetching user data:', error));
    }, []);

    if (!userData) {
        return <div>Loading...</div>;
    }

    // Function to render personal links as clickable links
    const renderPersonalLinks = () => {
        const linkKeys = ['personalLink1', 'personalLink2', 'personalLink3', 'personalLink4'];
        return linkKeys.map(key => (
            userData[key] ? <div key={key} className="nav-link-callout">
                <a className="nav-link" href={userData[key]} target="_blank" rel="noopener noreferrer">{userData[key]}</a>
            </div> : null
        ));
    };

    return (
        <div className="profile">
            <table>
                <tbody>
                    <tr>
                        <td colSpan="2" className="header">
                            <h3>Profile Settings</h3>
                        </td>
                    </tr>
                    <tr>
                        <td className="label">Username (Permanent):</td>
                        <td>{userData.username}</td>
                    </tr>
                    <tr>
                        <td className="label">Personal Emoji:</td>
                        <td>{userData.personalEmoji && <span style={{ fontSize: '36px' }}>{userData.personalEmoji}</span>}</td>
                    </tr>
                    <tr>
                        <td className="label">Display Name:</td>
                        <td>{userData.displayname}</td>
                    </tr>
                    <tr>
                        <td className="label">Description:</td>
                        <td>{userData.description}</td>
                    </tr>
                    <tr>
                        <td className="label">Personal Links:</td>
                        <td>{renderPersonalLinks()}</td>
                    </tr>
                    <tr>
                        <td colSpan="2" className="header">
                            <h3>Private Info</h3>
                        </td>
                    </tr>
                    <tr>
                        <td>Email (Fake for Now, No Update Capability):</td>
                        <td>{userData.email}</td>
                    </tr>
                    <tr>
                        <td>API Key (Not Valid Yet):</td>
                        <td>{userData.apiKey}</td>
                    </tr>
                    <tr>
                        <td colSpan="2" className="header">
                            <h3>Financial Stats</h3>
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="2" className="account-balance">
                            Account Balance: ${userData.accountBalance.toFixed(2)}
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    );


}

export default Profile;
