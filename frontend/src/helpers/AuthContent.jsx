import { API_URL } from '../config';
import React, { createContext, useContext, useState } from 'react';

const AuthContext = createContext();

function UseAuth() {
    return useContext(AuthContext);
}

const AuthProvider = ({ children }) => {
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    // Add more state as needed

    const login = async (username, password) => {
        try {
            const response = await fetch(`${API_URL}/api/v0/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            if (response.ok) {
                const data = await response.json();
                // Set login state, store token, etc.
                setIsLoggedIn(true);
                console.log('Login successful', data);
                // Store token in localStorage or manage it as needed
                localStorage.setItem('token', data.token);
                // Redirect or perform additional actions
            } else {
                console.error('Login failed');
                // Handle failure (e.g., set error state, show message)
            }
        } catch (error) {
            console.error('Login error:', error);
            // Handle error (e.g., set error state, show message)
        }
    };

    const logout = () => {
        // Perform logout logic
        setIsLoggedIn(false);
        // Clear token from localStorage or manage as needed
        localStorage.removeItem('token');
    };

    return (
        <AuthContext.Provider value={{ isLoggedIn, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export { UseAuth, AuthProvider };