import React, { useState } from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white rounded focus:outline-none transition duration-300";

// Default color is primary-pink, when selected has border
const SiteButton = () => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light border-transparent";
    const selectedButtonStyle = "bg-primary-pink border-transparent";

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={() => setIsSelected(!isSelected)}
        >
            SELECT
        </button>
    );
};

export default SiteButton;