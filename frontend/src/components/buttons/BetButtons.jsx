import React from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2";
const yesButtonStyle = `bg-green-btn hover:bg-green-btn-hover`;
const yesButtonHoverStyle = `bg-green-btn-hover hover:bg-green-btn`;
const noButtonStyle = `bg-red-btn hover:bg-red-btn`;
const noButtonHoverStyle = `bg-red-btn-hover hover:bg-red-btn`;
const neutralButtonStyle = `bg-neutral-btn hover:bg-neutral-btn-hover`;
const neutralButtonHoverStyle = `bg-neutral-btn-hover hover:bg-neutral-btn`;

const BetYesButton = ({ isSelected }) => (
<button
    className={`${buttonBaseStyle} ${isSelected ? yesButtonStyle : yesButtonHoverStyle}`}
>
    YES
</button>
);

const BetNoButton = ({ isSelected }) => (
<button
    className={`${buttonBaseStyle} ${isSelected ? noButtonStyle : noButtonHoverStyle}`}
>
    NO
</button>
);

const ConfirmBetButton = ({ onClick, selectedDirection }) => {
    const getButtonStyle = () => {
        switch (selectedDirection) {
            case 'NO':
                return "bg-red-500 hover:bg-red-700";
            case 'YES':
                return "bg-green-500 hover:bg-green-700";
            default:
                return "bg-custom-gray-light";
        }
    };

    const buttonText = () => {
        switch (selectedDirection) {
            case 'NO':
                return "CONFIRM BET NO";
            case 'YES':
                return "CONFIRM BET YES";
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


export { BetYesButton, BetNoButton, ConfirmBetButton };
