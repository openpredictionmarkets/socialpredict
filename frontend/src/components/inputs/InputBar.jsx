import React from 'react';

const RegularInput = ({ value, onChange }) => {
    return (
        <input
        type="text"
        placeholder="Regular"
        value={value}
        onChange={onChange}
        className="w-full px-4 py-2 border-2 border-blue-500 rounded-md text-white bg-transparent focus:outline-none"
        />
    );
};

const SuccessInput = ({ value, onChange }) => {
    return (
    <div className="flex items-center border-2 border-green-500 bg-transparent rounded-md">
        <input
        type="text"
        placeholder="Success"
        value={value}
        onChange={onChange}
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
        <span className="h-5 w-5 text-green-500 mr-2">✓</span>
    </div>
    );
};

const ErrorInput = ({ value, onChange }) => {
    return (
    <div className="flex items-center border-2 border-red-500 bg-transparent rounded-md">
        <input
        type="text"
        placeholder="Error Input"
        value={value}
        onChange={onChange}
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
        <span className="h-5 w-5 text-red-500 mr-2">✗</span>
    </div>
    );
};

const PersonInput = ({ value, onChange }) => {
    return (
    <div className="flex items-center border-2 border-blue-500 bg-transparent rounded-md">
        <span className="h-5 w-5 text-blue-500 ml-2">👤</span>
        <input
        type="text"
        placeholder="Username"
        value={value}
        onChange={onChange}
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
    </div>
    );
};

const LockInput = ({ value, onChange }) => {
    return (
    <div className="flex items-center border-2 border-blue-500 bg-transparent rounded-md">
        <span className="h-5 w-5 text-blue-500 ml-2">🔒</span>
        <input
        type="password"
        placeholder="Password"
        value={value}
        onChange={onChange}
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
    </div>
    );
};

export { RegularInput, SuccessInput, ErrorInput, PersonInput, LockInput };