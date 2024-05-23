import { API_URL } from '../../../config';
import React, { useState } from 'react';
import SiteButton from '../../buttons/SiteButtons';
import { RegularInput } from '../../inputs/InputBar'

function AdminAddUser() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [copied, setCopied] = useState(false);

    const handleUsernameChange = (event) => {
        setUsername(event.target.value);
    };

    const handleSubmit = async (event) => {
        event.preventDefault();
        setError('');
        try {
            const response = await fetch(`${API_URL}/api/v0/admin/createuser`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username })
            });
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            const data = await response.json();
            setPassword(data.password);
        } catch (err) {
            console.error('Failed to create user:', err);
            setError(err.message || 'Failed to create user');
        }
    };

    const handleCopyCredentials = () => {
        const credentials = `Username: ${username}\nPassword: ${password}`;
        navigator.clipboard.writeText(credentials).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);  // Notification timeout
        });
    };

    return (
        <div className="p-6 bg-primary-background shadow-md rounded-lg text-white">
            <h1 className="text-2xl font-bold mb-4">Create User</h1>
                <div className='Center-content-table'>
                    <form onSubmit={handleSubmit} className="space-y-6">
                        <RegularInput
                            type="text"
                            value={username}
                            onChange={handleUsernameChange}
                            placeholder="All lowercase letters and numbers"
                            required
                        />
                        <SiteButton type="submit">
                            Add User
                        </SiteButton>
                    </form>
                    {password && (
                        <div onClick={handleCopyCredentials} className="mt-4 p-4 bg-blue-500 text-white font-bold text-lg rounded-lg shadow-lg cursor-pointer flex justify-between items-center">
                            <div>
                                <p>Username: {username}</p>
                                <p>Password: {password}</p>
                            </div>
                            <div className="text-lg">
                                📋
                            </div>
                            {copied && <p className="text-green-500">COPIED!</p>}
                        </div>
                    )}
                    {error && <p className="error">{error}</p>}
                </div>
            </div>
    );
}

export default AdminAddUser;