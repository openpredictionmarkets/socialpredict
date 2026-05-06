import React, { useState } from 'react';
import { API_URL } from '../../../config';
import SiteButton from '../SiteButtons';
import EmojiPickerInput from '../../inputs/EmojiPicker';

const EmojiSelector = ({ currentEmoji = '', onSave }) => {
    const [selectedEmoji, setSelectedEmoji] = useState(currentEmoji || '');

    // JWT token retrieval
    const token = localStorage.getItem('token');
    if (!token) {
        alert('Authentication token not found. Please log in again.');
        throw new Error('Token not found');
    }

    const handleSave = async () => {
        const emoji = selectedEmoji.trim();

        if (!emoji) {
            alert('Please select an emoji before saving.');
            return;
        }

        try {
            const response = await fetch(`${API_URL}/v0/profilechange/emoji`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({ emoji }),
            });

            if (!response.ok) {
                throw new Error('Failed to change emoji');
            }

            await response.json();
            if (onSave) onSave(emoji);
        } catch (error) {
            console.error('Error changing emoji:', error);
        }
    };


    return (
        <div className="w-80 max-w-[80vw]">
            <EmojiPickerInput
                type="text"
                value={selectedEmoji}
                onChange={(event) => setSelectedEmoji(event.target.value)}
                placeholder="Choose an emoji"
                maxLength={20}
                replaceValueOnEmojiSelect
                className="w-full bg-gray-700 border border-gray-600 text-white px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div className="mt-4 px-4 py-2">
                <SiteButton onClick={handleSave}>
                    SAVE
                </SiteButton>
            </div>
        </div>
    );
};

export default EmojiSelector;
