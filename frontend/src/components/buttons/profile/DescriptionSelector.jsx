import { API_URL } from '../../../config';
import React, { useState } from 'react';
import SiteButton from '../SiteButtons';

const DescriptionSelector = ({ onSave }) => {
    const [description, setDescription] = useState('');

    const handleSave = async () => {
        if (!description.trim()) {
            alert('Please enter a description before saving.');
            return;
        }
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/profilechange/description`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({ description }),
            });
            const responseData = await response.json();
            if (response.ok) {
                console.log('Description updated successfully:', responseData);
                onSave(description);
            } else {
                throw new Error('Failed to update description');
            }
        } catch (error) {
            console.error('Error updating description:', error);
        }
    };

    return (
        <div>
            <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Enter new description..."
                className="mb-4 px-2 py-1 border rounded text-black"
            />
            <SiteButton onClick={handleSave}>Save Description</SiteButton>
        </div>
    );
};

export default DescriptionSelector;
