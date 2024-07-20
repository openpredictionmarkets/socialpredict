import React, { useState } from 'react';
import ReactDOM from 'react-dom';
import { useHistory } from 'react-router-dom';
import { PersonInput, LockInput } from '../../inputs/InputBar';
import SiteButton from '../../buttons/SiteButtons';
import { useAuth } from '../../../helpers/AuthContent';

const LoginModal = ({ isOpen, onClose, onLogin, redirectAfterLogin }) => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const history = useHistory();
    const { login, changePasswordNeeded } = useAuth();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        try {
            const loginSuccess = await onLogin(username, password);
            if (loginSuccess) {
                onClose();
                history.push(redirectAfterLogin);
                window.location.reload(); // this is a hack which goes against the principle of reloading the state
            } else {
                console.error('Login failed:', response.status, await response.text());
                setError('Error logging in.');
            }
        } catch (loginError) {
            console.error('Login error:', loginError);
            setError('An error occurred during login. Please try again.');
        }
    };

    if (!isOpen) return null;

    return ReactDOM.createPortal(
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
            <div className="relative bg-blue-900 p-6 rounded-lg text-white max-w-sm mx-auto">
                <h2 className="text-xl mb-4">Login</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <PersonInput value={username} onChange={(e) => {
                        setUsername(e.target.value);
                    }} />

                    <LockInput value={password} onChange={(e) => {
                        setPassword(e.target.value);
                    }} />
                    {error && <div className='error-message'>{error}</div>}
                    <div className="flex items-center justify-between">
                        <SiteButton type="submit">Login</SiteButton>
                    </div>
                </form>
                <button className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white" onClick={onClose}>
                    ✕
                </button>
            </div>
        </div>,
        document.getElementById('modal-root')
    );
};

export default LoginModal;