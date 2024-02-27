import React from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2";
// Updated to use Tailwind's custom colors
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

const ResolveButton = ({ isSelected }) => (
<button
    className={`${buttonBaseStyle} ${isSelected ? neutralButtonStyle : neutralButtonHoverStyle}`}
>
    RESOLVE
</button>
);

export { BetYesButton, BetNoButton, ResolveButton };
