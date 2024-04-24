import React from 'react';

const RegularInputBox = ({ value, onChange }) => {
    return (
        <textarea
            type="text"
            value={value}
            onChange={onChange}
            className="w-full h-64 px-4 py-2 border-2 border-blue-500 rounded-md text-white bg-transparent focus:outline-none resize-none"
        />
    );
};

export default RegularInputBox;