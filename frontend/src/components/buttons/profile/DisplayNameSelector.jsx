import { API_URL } from '../../../config';
import React, { useState } from 'react';
import SiteButton from '../SiteButtons';

const DisplayNameSelector = ({ onSave }) => {
    const [displayName, setDisplayName] = useState('');

    const handleSave = async () => {
        if (!displayName.trim()) {
            alert('Please enter a display name before saving.');
            return;
        }
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/profilechange/displayname`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({ displayName }),
            });
            const responseData = await response.json();
            if (response.ok) {
                console.log('Display name updated successfully:', responseData);
                onSave(displayName);
            } else {
                throw new Error('Failed to update display name');
            }
        } catch (error) {
            console.error('Error updating display name:', error);
        }
    };

    return (
        <div>
            <input
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="Enter new display name..."
                className="mb-4 px-2 py-1 border rounded text-black"
            />
            <SiteButton onClick={handleSave}>Save Display Name</SiteButton>
        </div>
    );
};

export default DisplayNameSelector;
