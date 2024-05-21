import React, { useState } from 'react';
import { buttonBaseStyle } from './BaseButton';

const SiteButton = ({ onClick, children }) => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light border-transparent";
    const selectedButtonStyle = "bg-primary-pink border-transparent";

    const handleClick = () => {
        setIsSelected(!isSelected);
        if (onClick) { // Only call onClick if it is provided
            onClick();
        }
    };

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={handleClick}
        >
            {children || 'SELECT'}
        </button>
    );
};

export default SiteButton;
