import React, { useState } from 'react';
import { PersonInput, LockInput } from '../inputs/InputBar';
import SiteButton from '../buttons/SiteButtons';

const LoginModal = ({ isOpen, onClose }) => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');

    if (!isOpen) return null;

    const handleSubmit = (e) => {
        e.preventDefault();
        // Handle the login logic here
        console.log('Username:', username, 'Password:', password);
        onClose(); // Close the modal on successful login
    };

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
            {/* Set the position of this div to relative to make the close button position relative to this container */}
            <div className="relative bg-blue-900 p-6 rounded-lg text-white max-w-sm mx-auto">
                <h2 className="text-xl mb-4">Login</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <PersonInput value={username} onChange={(e) => setUsername(e.target.value)} />
                    <LockInput value={password} onChange={(e) => setPassword(e.target.value)} />
                    <div className="flex items-center justify-between">
                        <SiteButton type="submit">Submit</SiteButton>
                    </div>
                </form>
                {/* This button is now positioned absolutely within its parent's relative position */}
                <button
                    className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white"
                    onClick={onClose}
                >
                    ✕
                </button>
            </div>
        </div>
    );
};

export default LoginModal;
