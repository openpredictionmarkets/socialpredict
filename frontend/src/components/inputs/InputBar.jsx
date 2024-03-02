import React from 'react';
// import { CheckIcon, XIcon, UserIcon } from '@heroicons/react/solid';

// Regular Input
const RegularInput = () => {
return (
    <input
    type="text"
    placeholder="Regular"
    className="w-full px-4 py-2 border-2 border-blue-500 rounded-md text-white bg-transparent focus:outline-none"
    />
);
};

const SuccessInput = () => {
    return (
    <div className="flex items-center border-2 border-green-500 bg-transparent rounded-md">
        <input
        type="text"
        placeholder="Success"
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
        <span className="h-5 w-5 text-green-500 mr-2">✓</span>
    </div>
    );
};

const ErrorInput = () => {
    return (
    <div className="flex items-center border-2 border-red-500 bg-transparent rounded-md">
        <input
        type="text"
        placeholder="Error Input"
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
        <span className="h-5 w-5 text-red-500 mr-2">✗</span>
    </div>
    );
};

const PersonInput = () => {
    return (
    <div className="flex items-center border-2 border-blue-500 bg-transparent rounded-md">
        <span className="h-5 w-5 text-blue-500 ml-2">👤</span>
        <input
        type="text"
        placeholder="Username"
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
    </div>
    );
};

const LockInput = () => {
    return (
    <div className="flex items-center border-2 border-blue-500 bg-transparent rounded-md">
        <span className="h-5 w-5 text-blue-500 ml-2">🔒</span>
        <input
        type="password"
        placeholder="Password"
        className="flex-1 px-4 py-2 rounded-md text-white bg-transparent focus:outline-none"
        />
    </div>
    );
};

export { RegularInput, SuccessInput, ErrorInput, PersonInput, LockInput };