import React from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2";
// Updated to use Tailwind's custom colors
const yesButtonStyle = "bg-green-btn hover:bg-green-btn-hover focus:ring-green-btn-border-hover";
const yesButtonHoverStyle = "bg-green-btn-hover hover:bg-green-btn focus:ring-green-btn-border-default";
const noButtonStyle = `bg-red-500 hover:bg-red-400 focus:ring-red-400`;
const noButtonHoverStyle = `bg-red-600 hover:bg-red-500 focus:ring-red-500`;

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

export { BetYesButton, BetNoButton };
