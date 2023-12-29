// User.js
import React, { useState, useEffect } from 'react';
import './User.css'; // Assuming you have a CSS file for styling
import { Link } from 'react-router-dom';

const User = ({ match }) => {
    const [userData, setUserData] = useState(null);
    const username = match.params.username;

    useEffect(() => {
        fetch(`https://brierfoxforecast.ngrok.app/api/v0/userinfo/${username}`)
            .then(response => response.json())
            .then(data => setUserData(data))
            .catch(error => console.error('Error fetching user data:', error));
    }, [username]);

    if (!userData) {
        return <div>Loading...</div>;
    }

    // Function to render personal links as clickable links
    const renderPersonalLinks = () => {
        const linkKeys = ['personalink1', 'personalink2', 'personalink3', 'personalink4'];
        return linkKeys.map(key => {
            if (userData[key]) {
                return (
                    <div key={key} className="nav-link-callout">
                        <a className="nav-link" href={userData[key]} target="_blank" rel="noopener noreferrer">{userData[key]}</a>
                    </div>
                );
            }
            return null;
        });
    };

    return (
        <div>
        <div className="user-public">
            <table>
                <tbody>
                    <tr>
                        <td className="emoji">
                            {userData.personalEmoji && <span style={{ fontSize: '48px' }}>{userData.personalEmoji}</span>}
                        </td>
                        <td className="display-name">
                            <h2>{userData.displayname}</h2>
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="2">
                            <Link to={`/user/${username}`}>@{username}</Link>
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="2" className="description">
                            <p>{userData.description}</p>
                        </td>
                    </tr>
                    <tr>
                        <td className="navlink" colSpan="2">
                            {renderPersonalLinks()}
                            </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div className="user-public">
            <table>
                <tbody>
                    <tr>
                        <td className="display-name">
                            <h4>Financials</h4>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <p>Account Balance: {userData.accountBalance.toFixed(2)}</p>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        </div>
    );
};

export default User;
