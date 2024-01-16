// EmojiSelector.js
import React, { useState } from 'react';

const EmojiSelector = ({ onSelectEmoji }) => {
    const emojis = ['😀', '😃', '😄', '😁', '😆', '😅', '😂', '🤣', '😊', '😇'];

    const handleEmojiSelect = (emoji) => {
      onSelectEmoji(emoji);
    };

    return (
      <div>
        {emojis.map((emoji) => (
          <button key={emoji} onClick={() => handleEmojiSelect(emoji)}>
            {emoji}
          </button>
        ))}
      </div>
    );
  };


export default EmojiSelector;
