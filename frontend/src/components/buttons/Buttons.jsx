import React, { useState } from 'react';

const buttonBaseStyle = `w-full px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2`;
const yesButtonStyle = `bg-green-500 hover:bg-green-400 focus:ring-green-400`;
const noButtonStyle = `bg-red-500 hover:bg-red-400 focus:ring-red-400`;
const yesButtonHoverStyle = `bg-green-600 hover:bg-green-500 focus:ring-green-500`;
const noButtonHoverStyle = `bg-red-600 hover:bg-red-500 focus:ring-red-500`;

const BetYesButton = ({ isSelected }) => (
  <button
    className={`${buttonBaseStyle} ${isSelected ? yesButtonStyle : yesButtonHoverStyle}`}
    style={{
      boxShadow: '0 4px 6px rgba(50, 50, 93, 0.11), 0 1px 3px rgba(0, 0, 0, 0.08)',
      backgroundColor: isSelected ? '#00bf9a' : '#00cca4',
      borderColor: isSelected ? '#00bf9a' : '#00cca4'
    }}
    onMouseEnter={(e) => {
      e.target.style.backgroundColor = isSelected ? '#00cca4' : '#00f2c3';
      e.target.style.borderColor = isSelected ? '#00cca4' : '#00f2c3';
    }}
    onMouseLeave={(e) => {
      e.target.style.backgroundColor = isSelected ? '#00bf9a' : '#00cca4';
      e.target.style.borderColor = isSelected ? '#00bf9a' : '#00cca4';
    }}
  >
    YES
  </button>
);

const BetNoButton = ({ isSelected }) => (
  <button
    className={`${buttonBaseStyle} ${isSelected ? noButtonStyle : noButtonHoverStyle}`}
    style={{
      boxShadow: '0 4px 6px rgba(50, 50, 93, 0.11), 0 1px 3px rgba(0, 0, 0, 0.08)',
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
