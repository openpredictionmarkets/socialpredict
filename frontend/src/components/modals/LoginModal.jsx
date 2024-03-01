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
    <div className="bg-blue-900 p-6 rounded-lg text-white max-w-sm mx-auto">
        <h2 className="text-xl mb-4">Sign in with</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
        <PersonInput value={username} onChange={(e) => setUsername(e.target.value)} />
        <LockInput value={password} onChange={(e) => setPassword(e.target.value)} />
        <div className="flex items-center justify-between">
            <label className="flex items-center">
            <input type="checkbox" className="form-checkbox" />
            <span className="ml-2">Remember me!</span>
            </label>
            <SiteButton type="submit" onClick={handleSubmit} />
        </div>
        </form>
        <button
        className="absolute top-4 right-4 text-gray-400 hover:text-white"
        onClick={onClose}
        >
        ✕
        </button>
    </div>
    </div>
);
};

export default LoginModal;