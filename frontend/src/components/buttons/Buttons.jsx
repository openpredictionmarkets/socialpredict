import React, { useState } from 'react';
import Colors from '../../config/Colors';

const buttonBaseStyle = "w-full px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2";
// Updated to use Tailwind's custom colors
const yesButtonStyle = "bg-green-btn hover:bg-green-btn-hover focus:ring-green-btn-border-hover";
const yesButtonHoverStyle = "bg-green-btn-hover hover:bg-green-btn focus:ring-green-btn-border-default";
const boxShadowStyle = "shadow-xl"; // Example of using Tailwind's box shadow utility instead of inline style

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
    style={{
        boxShadow: boxShadowStyle,
        backgroundColor: isSelected ? '#fd5d93' : '#ec250d',
        borderColor: isSelected ? '#fd5d93' : '#ec250d'
    }}
    onMouseEnter={(e) => {
        e.target.style.backgroundColor = isSelected ? '#ec250d' : '#fd5d93';
        e.target.style.borderColor = isSelected ? '#ec250d' : '#fd5d93';
    }}
    onMouseLeave={(e) => {
        e.target.style.backgroundColor = isSelected ? '#fd5d93' : '#ec250d';
        e.target.style.borderColor = isSelected ? '#fd5d93' : '#ec250d';
    }}
>
    NO
</button>
);

export { BetYesButton, BetNoButton };
