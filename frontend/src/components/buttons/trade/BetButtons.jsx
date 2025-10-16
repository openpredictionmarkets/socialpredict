import React, { useState } from 'react';
import { buttonBaseStyle } from '../BaseButton';
import { NumberInput } from '../../inputs/InputBar';

// Toggle buttons between initial and selected states
const BetButton = ({ onClick }) => {
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
            TRADE
        </button>
    );
};

const BetNoButton = ({ onClick, label = "NO" }) => {
    return (
        <button
        className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-red-btn`}
            onClick={onClick}
        >
            {label}
        </button>
    );
};

const BetYesButton = ({ onClick, label = "YES" }) => {
    return (
        <button
        className={`${buttonBaseStyle} bg-custom-gray-light hover:bg-green-btn`}
            onClick={onClick}
        >
            {label}
        </button>
    );
};

const BetInputAmount = ({ value, onChange }) => {
    return (
        <NumberInput
            value={value}
            onChange={onChange}
        />
    );
};

const ConfirmBetButton = ({ onClick, selectedDirection, yesLabel = "YES", noLabel = "NO" }) => {
    const getButtonStyle = () => {
        switch (selectedDirection) {
            case 'NO':
                return "bg-red-btn hover:bg-red-btn";
            case 'YES':
                return "bg-green-btn hover:bg-green-btn";
            default:
                return "bg-custom-gray-light";
        }
    };

    const buttonText = () => {
        switch (selectedDirection) {
            case 'NO':
                return `CONFIRM PURCHASE OF: ${noLabel}`;
            case 'YES':
                return `CONFIRM PURCHASE OF: ${yesLabel}`;
            default:
                return "CONFIRM PURCHASE";
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



export { BetButton, BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton };
