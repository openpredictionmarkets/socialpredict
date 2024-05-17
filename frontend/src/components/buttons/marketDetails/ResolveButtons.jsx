import React, { useState } from 'react';
import { buttonBaseStyle } from '../BaseButton';

// Toggle buttons between initial and selected states
const ResolveButton = ({ onClick }) => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light";
    const selectedButtonStyle = "bg-neutral-btn";
    const buttonBaseStyle = "w-full px-4 py-2 text-white border rounded focus:outline-none";

    const handleClick = () => {
        setIsSelected(!isSelected);
        onClick && onClick();
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
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-red-btn`}
            onClick={onClick}
        >
            RESOLVE NO
        </button>
    );
};

const SelectYesButton = ({ onClick }) => {
    return (
        <button
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-green-btn`}
            onClick={onClick}
        >
            RESOLVE YES
        </button>
    );
};


const ConfirmResolveButton = ({ onClick, selectedResolution }) => {
    const getButtonStyle = () => {
        switch (selectedResolution) {
            case 'NO':
                return "bg-red-btn hover:bg-red-btn";
            case 'YES':
                return "bg-green-btn hover:bg-green-btn";
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
            onClick={onClick}
        >
            {buttonText()}
        </button>
    );
};

export { ResolveButton, SelectNoButton, SelectYesButton, ConfirmResolveButton };
