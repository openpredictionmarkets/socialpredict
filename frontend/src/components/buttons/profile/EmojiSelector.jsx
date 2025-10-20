import React, { useState, useEffect } from 'react';
import { API_URL } from '../../../config';
import ProfileEditButton from './ProfileButtons';
import SiteButton from '../SiteButtons';
import { emojis } from './Emojis';

const EmojiSelector = ({ onSave }) => {
    const [selectedEmoji, setSelectedEmoji] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [filteredEmojis, setFilteredEmojis] = useState([]);

    // JWT token retrieval
    const token = localStorage.getItem('token');
    if (!token) {
        alert('Authentication token not found. Please log in again.');
        throw new Error('Token not found');
    }

    useEffect(() => {
        const regex = new RegExp(searchTerm, 'i');
        const numberOfEmojisShown = 20;
        setFilteredEmojis(emojis.filter(emoji => regex.test(emoji.name)).slice(0, numberOfEmojisShown));
    }, [searchTerm]);

    useEffect(() => {
        console.log('Selected emoji updated:', selectedEmoji);
    }, [selectedEmoji]);

    const handleEmojiClick = (emoji) => {
        console.log('Emoji clicked:', emoji);
        setSelectedEmoji(emoji);
    };

    const handleSearchChange = (event) => {
        setSearchTerm(event.target.value);
    };

    const handleSave = async () => {
        if (!selectedEmoji) {
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
                body: JSON.stringify({ emoji: selectedEmoji.symbol }), // Send only the symbol
            });

            console.log('Response status:', response.status); // Log response status
            if (!response.ok) {
                throw new Error('Failed to change emoji');
            }
            const responseData = await response.json();
            console.log('Response data:', responseData); // Log response data

            console.log('Emoji changed successfully');
            if (onSave) onSave(selectedEmoji.symbol);
        } catch (error) {
            console.error('Error changing emoji:', error);
        }
    };


    return (
        <div>
            <input
                type="text"
                value={searchTerm}
                onChange={handleSearchChange}
                placeholder="Search emojis..."
                className="mb-4 px-2 py-1 border rounded text-black"
            />
            {searchTerm && (
                <div className="grid grid-cols-4 gap-1 max-h-56 overflow-auto">
                    {filteredEmojis.map((emoji) => (
                        <ProfileEditButton
                            key={emoji.name}
                            onClick={() => handleEmojiClick(emoji)}
                            isSelected={selectedEmoji && selectedEmoji.name === emoji.name}
                        >
                            {emoji.symbol}
                        </ProfileEditButton>
                    ))}
                </div>
            )}
            <div className="mt-4 px-4 py-2">
                <SiteButton onClick={handleSave}>
                    SAVE
                </SiteButton>
            </div>
        </div>
    );
};

export default EmojiSelector;
