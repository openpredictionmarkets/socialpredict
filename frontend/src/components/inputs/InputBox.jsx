import React from 'react';

const RegularInputBox = () => {
    return (
        <textarea
            placeholder="Regular"
            className="w-full h-64 px-4 py-2 border-2 border-blue-500 rounded-md text-white bg-transparent focus:outline-none resize-none"
        />
    );
};

export default RegularInputBox;