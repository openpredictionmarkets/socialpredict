import React, { useState } from 'react';
import { buttonBaseStyle } from '../BaseButton';
import { NumberInput } from '../../inputs/InputBar';


const SharesBadgeSimple = ({ type, count, label }) => {

    const backgrounds = {
        YES: 'linear-gradient(to right, #054A29, #FFC107)', // green to yellow gradient
        NO: 'linear-gradient(to right, #D00000, #FFC107)',  // red to yellow gradient
        // YES: `linear-gradient(to right, theme('colors.green-btn'), theme('colors.gold-btn')`,
        // NO: `linear-gradient(to right, theme('colors.red-btn'), theme('colors.gold-btn')`,
    };

    // Use custom label if provided, otherwise fall back to type
    const displayLabel = label || type;

    return (
        <div
            className="cursor-default text-white px-4 py-2 rounded-full shadow-md"
            style={{ background: backgrounds[type] }}
        >
            {displayLabel}: {count} 🪙
        </div>
    );
};

const SharesBadge = ({ type, count, label }) => {
    const colors = {
        YES: '#054A29',
        NO: '#D00000',
        GOLD: '#FFC107',
        BEIGE: '#F9D3A5'
    };

    const backgroundStyle = {
        backgroundImage: `linear-gradient(to right, ${colors[type]}, ${colors.BEIGE})`,
        color: 'black',
        borderColor: `${colors.GOLD}`,
    };

    // Use custom label if provided, otherwise fall back to type
    const displayLabel = label || type;

    return (
        <div
            className="cursor-default px-4 py-2 font-bold text-lg"
            style={{
                ...backgroundStyle,
                border: `3px solid ${colors.GOLD}`,
                borderRadius: '12px',
                position: 'relative',
            }}
        >
            {displayLabel}: {count} 🪙
        </div>
    );
};

const SaleInputAmount = ({ value, onChange, max }) => {
    return (
        <NumberInput
            value={value}
            onChange={onChange}
            max={max}
        />
    );
};

const ConfirmSaleButton = ({ onClick, selectedDirection }) => {
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
                return "CONFIRM SALE";
            case 'YES':
                return "CONFIRM SALE";
            default:
                return "CONFIRM SALE";
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

export { SharesBadge, SharesBadgeSimple, SaleInputAmount, ConfirmSaleButton, }