import React, { useState } from 'react';

function AdminDashboard() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

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

    return (
        <div className='App'>
            <div className='Center-content'>
                <div className='Center-content-header'>
                    <h1>Create User</h1>
                </div>
                <div className='Center-content-table'>
                    <form onSubmit={handleSubmit}>
                        <input
                            type="text"
                            value={username}
                            onChange={handleUsernameChange}
                            placeholder="Enter username"
                            required
                        />
                        <button type="submit">Add</button>
                    </form>
                    {password && (
                        <div>
                            <h2>Generated Password: {password}</h2>
                            <p>Copy and provide this password to the user.</p>
                        </div>
                    )}
                    {error && <p className="error">{error}</p>}
                </div>
            </div>
        </div>
    );
}

export default AdminDashboard;
