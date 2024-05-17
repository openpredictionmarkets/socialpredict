import React, { useState } from 'react';
import { API_URL } from '../../../config';
import PersonalEmojiButton from './ProfileButtons';
import SiteButton from '../SiteButtons';
import { emojis } from './Emojis';

const EmojiSelector = ({ username, onSave }) => {
    const [selectedEmoji, setSelectedEmoji] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [filteredEmojis, setFilteredEmojis] = useState([]);

    const handleEmojiClick = (emoji) => {
        setSelectedEmoji(emoji);
    };

    const handleSearchChange = (event) => {
        const { value } = event.target;
        setSearchTerm(value);
        if (value) {
            const regex = new RegExp(value, 'i');
            setFilteredEmojis(emojis.filter(emoji => regex.test(emoji.name)));
        } else {
            setFilteredEmojis([]);
        }
    };

    /*
    const handleSave = async () => {
        try {
            const response = await fetch(`${API_URL}/v0/api/emojichange`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, emoji: selectedEmoji }),
            });
            if (!response.ok) {
                throw new Error('Failed to change emoji');
            }
            console.log('Emoji changed successfully');
            if (onSave) onSave(selectedEmoji);
        } catch (error) {
            console.error('Error changing emoji:', error);
        }
    };
    */

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
                <div className="flex flex-wrap space-x-2 space-y-2">
                    {filteredEmojis.map((emoji) => (
                        <button
                            key={emoji.name}
                            onClick={() => handleEmojiClick(emoji.symbol)}
                            className="text-lg p-1"
                        >
                            {emoji.symbol}
                        </button>
                    ))}
                </div>
            )}
            <div className="mt-4 px-4 py-2">
                <SiteButton onClick={() => onSave(selectedEmoji)}>
                    SAVE
                </SiteButton>
            </div>
        </div>
    );
};

export default EmojiSelector;
