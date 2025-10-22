import React, { useState, useRef, useEffect } from 'react';
import EmojiPicker from 'emoji-picker-react';

const EmojiPickerInput = ({ 
  value, 
  onChange, 
  placeholder, 
  className = '', 
  type = 'text',
  maxLength,
  ...props 
}) => {
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [cursorPosition, setCursorPosition] = useState(0);
  const inputRef = useRef(null);
  const pickerRef = useRef(null);

  // Handle clicking outside to close emoji picker
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (
        pickerRef.current && 
        !pickerRef.current.contains(event.target) &&
        !event.target.closest('.emoji-picker-button')
      ) {
        setShowEmojiPicker(false);
      }
    };

    if (showEmojiPicker) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showEmojiPicker]);

  // Track cursor position
  const handleInputChange = (e) => {
    setCursorPosition(e.target.selectionStart);
    onChange(e);
  };

  const handleInputClick = (e) => {
    setCursorPosition(e.target.selectionStart);
  };

  const handleInputKeyUp = (e) => {
    setCursorPosition(e.target.selectionStart);
  };

  // Handle emoji selection
  const onEmojiClick = (emojiData) => {
    const emoji = emojiData.emoji;
    const newValue = value.slice(0, cursorPosition) + emoji + value.slice(cursorPosition);
    
    // Create synthetic event object
    const syntheticEvent = {
      target: {
        value: newValue,
        selectionStart: cursorPosition + emoji.length
      }
    };
    
    onChange(syntheticEvent);
    
    // Update cursor position and focus back to input
    setTimeout(() => {
      if (inputRef.current) {
        inputRef.current.focus();
        inputRef.current.setSelectionRange(
          cursorPosition + emoji.length, 
          cursorPosition + emoji.length
        );
        setCursorPosition(cursorPosition + emoji.length);
      }
    }, 10);
    
    setShowEmojiPicker(false);
  };

  const toggleEmojiPicker = () => {
    setShowEmojiPicker(!showEmojiPicker);
    if (!showEmojiPicker && inputRef.current) {
      setCursorPosition(inputRef.current.selectionStart || value.length);
    }
  };

  const inputClasses = `
    ${className}
    pr-10
  `.trim();

  return (
    <div className="relative">
      <div className="relative">
        {type === 'textarea' ? (
          <textarea
            ref={inputRef}
            value={value}
            onChange={handleInputChange}
            onClick={handleInputClick}
            onKeyUp={handleInputKeyUp}
            placeholder={placeholder}
            maxLength={maxLength}
            className={inputClasses}
            {...props}
          />
        ) : (
          <input
            ref={inputRef}
            type={type}
            value={value}
            onChange={handleInputChange}
            onClick={handleInputClick}
            onKeyUp={handleInputKeyUp}
            placeholder={placeholder}
            maxLength={maxLength}
            className={inputClasses}
            {...props}
          />
        )}
        
        <button
          type="button"
          onClick={toggleEmojiPicker}
          className="emoji-picker-button absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200 transition-colors p-1 rounded"
          title="Add emoji"
        >
          😀
        </button>
      </div>

      {showEmojiPicker && (
        <div 
          ref={pickerRef}
          className="absolute z-50 mt-2 right-0"
          style={{ 
            maxWidth: '100vw',
            transform: 'translateX(0)'
          }}
        >
          <EmojiPicker
            onEmojiClick={onEmojiClick}
            width={320}
            height={400}
            theme="dark"
            previewConfig={{
              showPreview: false
            }}
            searchDisabled={false}
            skinTonesDisabled={true}
            categories={[
              'smileys_people',
              'animals_nature',
              'food_drink',
              'activities',
              'travel_places',
              'objects',
              'symbols',
              'flags'
            ]}
          />
        </div>
      )}
    </div>
  );
};

export default EmojiPickerInput;