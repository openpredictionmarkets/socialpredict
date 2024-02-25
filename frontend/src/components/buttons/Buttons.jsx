import React from 'react';

const YesButton = ({ isSelected, onClick }) => (
<button
    className={`px-4 py-2 text-white border border-transparent rounded focus:outline-none focus:ring-2 focus:ring-offset-2 ${isSelected ? 'bg-green-400 hover:bg-green-300 focus:ring-green-300' : 'bg-green-500 hover:bg-green-400 focus:ring-green-400'}`}
    style={{
        boxShadow: '0 4px 6px rgba(50, 50, 93, 0.11), 0 1px 3px rgba(0, 0, 0, 0.08)',
        backgroundColor: isSelected ? '#00cca4' : '#00f2c3', // Adjust based on selection
        borderColor: isSelected ? '#00bf9a' : '#00f2c3'
    }}
    onMouseEnter={(e) => {
        e.target.style.backgroundColor = isSelected ? '#00bf9a' : '#00cca4';
        e.target.style.borderColor = isSelected ? '#00b290' : '#00bf9a';
    }}
    onMouseLeave={(e) => {
        e.target.style.backgroundColor = isSelected ? '#00cca4' : '#00f2c3';
        e.target.style.borderColor = isSelected ? '#00bf9a' : '#00f2c3';
    }}
    onClick={onClick}
>
    YES
</button>
);

export default YesButton;