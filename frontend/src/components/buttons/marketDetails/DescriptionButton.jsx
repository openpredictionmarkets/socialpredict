import React, { useState } from 'react';
import { buttonBaseStyle } from '../BaseButton';

const DescriptionButton = ({ onClick, children }) => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light border-transparent";
    const selectedButtonStyle = "bg-primary-pink border-transparent";

    const handleButtonClick = () => {
        setIsSelected(!isSelected);
        if (onClick) onClick();
    };

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={handleButtonClick}
        >
            {children || 'SELECT'}
        </button>
    );
};

export default DescriptionButton;