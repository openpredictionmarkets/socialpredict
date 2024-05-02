import React, { useState } from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white border rounded focus:outline-none";

// Toggle buttons between initial and selected states
const ResolveButton = ({ onClick }) => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light";
    const selectedButtonStyle = "bg-neutral-btn";
    const buttonBaseStyle = "w-full px-4 py-2 text-white border rounded focus:outline-none";

    const handleClick = () => {
        setIsSelected(!isSelected); // Toggle the internal selected state
        onClick && onClick(); // Call the external onClick handler
    };

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle} min-w-32 text-xs sm:text-sm md:text-base`}
            onClick={handleClick}
        >
            RESOLVE
        </button>
    );
};

const SelectNoButton = ({ onClick }) => {
    return (
        <button
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-red-500`}
            onClick={onClick}
        >
            NO
        </button>
    );
};

const SelectYesButton = ({ onClick }) => {
    return (
        <button
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-green-500`}
            onClick={onClick}
        >
            YES
        </button>
    );
};


const ConfirmResolveButton = ({ selectedResolution }) => {
    const getButtonStyle = () => {
        switch (selectedResolution) {
            case 'NO':
                return "bg-red-500 hover:bg-red-700";
            case 'YES':
                return "bg-green-500 hover:bg-green-700";
            default:
                return "bg-custom-gray-light";
        }
    };

    const buttonText = () => {
        switch (selectedResolution) {
            case 'NO':
                return "CONFIRM RESOLVE NO";
            case 'YES':
                return "CONFIRM RESOLVE YES";
            default:
                return "CONFIRM";
        }
    };

    return (
        <button
            className={`w-full px-4 py-2 text-white border rounded focus:outline-none ${getButtonStyle()}`}
        >
            {buttonText()}
        </button>
    );
};

export { ResolveButton, SelectNoButton, SelectYesButton, ConfirmResolveButton };
