import React, { useState, useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { PersonInput, LockInput } from '../inputs/InputBar';
import SiteButton from '../buttons/SiteButtons';
import UserContext from '../../helpers/UserContext'; // Adjust the path as necessary


const LoginModal = ({ isOpen, onClose, onLogin }) => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const history = useHistory();

    const { setUsername: setContextUsername } = useContext(UserContext);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (onLogin) {
            try {
                await onLogin(username, password); // Execute login logic
                setContextUsername(username); // Update username in context
                onClose(); // Close the modal on successful login
                history.push('/'); // Redirect to home
            } catch (loginError) {
                console.error('Login error:', loginError);
                setError('Incorrect login credentials'); // Set error message
            }
            } else {
            console.error('onLogin prop is not a function');
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
            <div className="relative bg-blue-900 p-6 rounded-lg text-white max-w-sm mx-auto">
                <h2 className="text-xl mb-4">Login</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <PersonInput value={username} onChange={(e) => setUsername(e.target.value)} />
                    <LockInput value={password} onChange={(e) => setPassword(e.target.value)} />
                    {error && <div className='error-message'>{error}</div>}
                    <div className="flex items-center justify-between">
                        <SiteButton type="submit">Login</SiteButton>
                    </div>
                </form>
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