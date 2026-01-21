import React, { useState } from 'react';
import { buttonBaseStyle } from '../BaseButton';

// Toggle buttons between initial and selected states
const ResolveButton = ({ onClick }) => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light";
    const selectedButtonStyle = "bg-neutral-btn";

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

const SelectNoButton = ({ onClick, label = "NO" }) => {
    return (
        <button
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-red-btn`}
            onClick={onClick}
        >
            RESOLVE {label}
        </button>
    );
};

const SelectYesButton = ({ onClick, label = "YES" }) => {
    return (
        <button
            className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-green-btn`}
            onClick={onClick}
        >
            RESOLVE {label}
        </button>
    );
};


const ConfirmResolveButton = ({ onClick, selectedResolution, yesLabel = "YES", noLabel = "NO" }) => {
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
                return `CONFIRM RESOLVE ${noLabel}`;
            case 'YES':
                return `CONFIRM RESOLVE ${yesLabel}`;
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
