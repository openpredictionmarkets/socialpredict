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

export const SiteButtonSmall = ({
    onClick,
    children = 'Action',
    type = 'button',
    disabled = false,
    className = '',
}) => {
    return (
        <button
            type={type}
            onClick={onClick}
            disabled={disabled}
            className={`${buttonBaseStyle} px-4 py-2 text-sm bg-primary-pink border-transparent disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
        >
            {children}
        </button>
    );
};

export default SiteButton;
