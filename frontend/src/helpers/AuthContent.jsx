import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext({
    username: null,
    setUsername: () => {},
    isLoggedIn: false
});

const useAuth = () => useContext(
    AuthContext
);

const AuthProvider = ({ children }) => {
    const [authState, setAuthState] = useState({
        isLoggedIn: false,
        token: localStorage.getItem('token'),
        username: localStorage.getItem('username'),
    });

    useEffect(() => {
        const token = localStorage.getItem('token');
        // Additional validation or expiry check can be performed here
        if (token) {
            setAuthState(prevState => ({
                ...prevState,
                isLoggedIn: true, // Assume token is still valid
                token: token,
                username: localStorage.getItem('username'),
            }));
        }
    }, []);

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
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', data.username);
                setAuthState({
                    isLoggedIn: true,
                    token: data.token,
                    username: data.username,
                });
                return true;
            } else {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    };

    const logout = () => {
        localStorage.clear();
        setAuthState({
            isLoggedIn: false,
            token: null,
            username: null,
        });
    };

    return (
        <AuthContext.Provider value={{ ...authState, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export { useAuth, AuthProvider, AuthContext };
