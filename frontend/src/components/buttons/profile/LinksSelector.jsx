import { API_URL } from '../../../config';
import React, { useState } from 'react';
import SiteButton from '../SiteButtons';

const LinksSelector = ({ onSave }) => {
    const [links, setLinks] = useState({ link1: '', link2: '', link3: '', link4: '' });

    const handleSave = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/api/v0/profilechange/links`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify(links),
            });
            const responseData = await response.json();
            if (response.ok) {
                console.log('Links updated successfully:', responseData);
                onSave(links);
            } else {
                throw new Error('Failed to update links');
            }
        } catch (error) {
            console.error('Error updating links:', error);
        }
    };

    return (
        <div>
            {Object.keys(links).map((key, index) => (
                <input
                    key={key}
                    type="text"
                    value={links[key]}
                    onChange={(e) => setLinks({ ...links, [key]: e.target.value })}
                    placeholder={`Enter link ${index + 1}...`}
                    className="mb-2 px-2 py-1 border rounded text-black"
                />
            ))}
            <SiteButton onClick={handleSave}>Save Links</SiteButton>
        </div>
    );
};

export default LinksSelector;
