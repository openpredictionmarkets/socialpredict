import React, { useState } from 'react';

export default function TradeCTA({ onClick, disabled }) {
  const [isSelected, setIsSelected] = useState(false);
  const initialButtonStyle = "bg-custom-gray-light";
  const selectedButtonStyle = "bg-neutral-btn";
  const buttonBaseStyle = "w-full px-4 py-2 text-white border rounded focus:outline-none";

  const handleClick = () => {
    setIsSelected(!isSelected);
    onClick && onClick();
  };

  return (
    <div
      className="md:hidden fixed inset-x-0 bottom-0 z-40 bg-primary-background/90 backdrop-blur p-3"
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 80px)' }}
      data-testid="mobile-trade-cta"
    >
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled}
        className={`${buttonBaseStyle} ${isSelected ? selectedButtonStyle : initialButtonStyle} min-w-32 text-xs sm:text-sm md:text-base disabled:opacity-50`}
      >
        TRADE
      </button>
    </div>
  );
}
