import React, { useState } from 'react';

const buttonBaseStyle = "w-full px-4 py-2 text-white border rounded focus:outline-none";

// Toggle buttons between initial and selected states
const ResolveButton = () => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light"; // Assuming this is your neutral initial state
    const selectedButtonStyle = "bg-neutral-btn"; // Assuming this is your selected state color

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={() => setIsSelected(!isSelected)}
        >
            RESOLVE
        </button>
    );
};

const ConfirmNoButton = () => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light";
    const selectedButtonStyle = "bg-red-btn";

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={() => setIsSelected(!isSelected)}
        >
            NO
        </button>
    );
};

const ConfirmYesButton = () => {
    const [isSelected, setIsSelected] = useState(false);
    const initialButtonStyle = "bg-custom-gray-light";
    const selectedButtonStyle = "bg-green-btn";

    return (
        <button
            className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle}`}
            onClick={() => setIsSelected(!isSelected)}
        >
            YES
        </button>
    );
};

export { ResolveButton, ConfirmNoButton, ConfirmYesButton };
