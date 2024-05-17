import React, { useState } from 'react';
import { buttonBaseStyle } from './BaseButton';


// Default color is primary-pink, when selected has border
const SiteButton = ({ children }) => { // Accept children as a prop
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light border-transparent";
    const selectedButtonStyle = "bg-primary-pink border-transparent";

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={() => setIsSelected(!isSelected)}
        >
            {children || 'SELECT'} {/* Use children if provided, otherwise default to 'SELECT' */}
        </button>
    );
};

export default SiteButton;