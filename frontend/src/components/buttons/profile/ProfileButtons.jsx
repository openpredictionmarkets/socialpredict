import React from 'react';

// Existing Emoji Button
const ProfileEditButton = ({ onClick, children, isSelected }) => {
    const initialButtonStyle = "bg-custom-gray-light border-transparent";
    const selectedButtonStyle = "bg-primary-pink border-transparent";

    return (
        <button
            className={`p-1 rounded-sm text-lg ${isSelected ? selectedButtonStyle : initialButtonStyle} flex items-center justify-center`}
            onClick={onClick}
        >
            {children}
        </button>
    );
};

export default ProfileEditButton;
