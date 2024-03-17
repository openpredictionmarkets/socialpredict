import { API_URL } from '../../config';
import React, { useState } from 'react';
import LoginModal from './LoginModal';

const LoginModalButton = () => {
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);

    const handleOpenModal = () => {
        setIsLoginModalOpen(true);
    };

    // Define the onLogin function here
    const onLogin = async (username, password) => {
        // Placeholder login logic - replace this with your actual login logic
        console.log('Attempting to log in with:', username, password);
        try {
            const response = await fetch(`${API_URL}/api/v0/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });
            if (response.ok) {
                await onLogin(username, password); // Let App.js handle login
                const data = await response.json();
                console.log('Login successful', data);
                setIsLoginModalOpen(false); // Close the modal on successful login
                setContextUsername(username); // Set username in context
                setIsLoggedIn(true);
                setUsername(username); // Set the username
                console.log('Logged in as:', username); // Log out the username
                const token = responseData.token;
                console.log('JWT Key:', token); // Log the JWT key
                localStorage.setItem('token', token);
                history.push('/'); // Redirect after successful login
            } else {
                // Handle login failure (e.g., show error message)
                console.error('Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
        }
    };

    return (
        <div>
            <button onClick={handleOpenModal}>Login</button>
            {isLoginModalOpen && <LoginModal isOpen={isLoginModalOpen} onClose={() => setIsLoginModalOpen(false)} onLogin={onLogin} />}
        </div>
    );
};

export default LoginModalButton;
