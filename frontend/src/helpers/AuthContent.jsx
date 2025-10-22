import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext({
    username: null,
    setUsername: () => {},
    isLoggedIn: false,
    usertype: null,
    changePasswordNeeded: true, // Default to true until login confirms otherwise
    login: () => {},
    logout: () => {},
});

const useAuth = () => useContext(
    AuthContext
);

const AuthProvider = ({ children }) => {
    const [authState, setAuthState] = useState({
        isLoggedIn: false,
        token: localStorage.getItem('token'),
        username: localStorage.getItem('username'),
        usertype: localStorage.getItem('usertype'),
        changePasswordNeeded: null  // Initialized as null
    });

    useEffect(() => {
        if (authState.isLoggedIn && authState.usertype) {
            // Redirect or perform other actions based on usertype
        }
    }, [authState.isLoggedIn, authState.usertype]);

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            setAuthState(prevState => ({
                ...prevState,
                isLoggedIn: true,
                token: token,
                username: localStorage.getItem('username'),
                usertype: localStorage.getItem('usertype'),
                // assume password change needed until shown not
                changePasswordNeeded: localStorage.getItem('changePasswordNeeded') === 'true',
            }));
        }
    }, []);

    const login = async (username, password) => {
        try {
            const response = await fetch(`${API_URL}/v0/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            // Read response as text first to handle both JSON and non-JSON responses
            const text = await response.text();
            let data = {};

            // Safely attempt to parse JSON
            try {
                data = JSON.parse(text);
            } catch (parseError) {
                // If JSON parsing fails, create a basic error object
                data = { error: text || 'Unknown error occurred' };
            }

            if (response.ok) {
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', data.username);
                localStorage.setItem('usertype', data.usertype);
                localStorage.setItem('changePasswordNeeded', data.mustChangePassword);
                setAuthState({
                    isLoggedIn: true,
                    token: data.token,
                    username: data.username,
                    usertype: data.usertype,
                    changePasswordNeeded: data.mustChangePassword,
                });
                return true;
            } else {
                // Create meaningful error message based on response
                const errorMessage = data.error || data.message || `HTTP ${response.status}: ${text}`;
                throw new Error(errorMessage);
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
            usertype: null,
            changePasswordNeeded: null,
        });
    };

    return (
        <AuthContext.Provider value={{ ...authState, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export { useAuth, AuthProvider, AuthContext };
