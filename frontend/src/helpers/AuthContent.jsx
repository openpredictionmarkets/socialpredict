import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext({
    username: null,
    setUsername: () => {},
    isLoggedIn: false,
    usertype: null,
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
    });

    useEffect(() => {
        if (authState.isLoggedIn && authState.usertype) {
            console.log("Logged in as:", authState.usertype);
            // Redirect or perform other actions based on usertype
        }
    }, [authState.isLoggedIn, authState.usertype]);

    useEffect(() => {
        const token = localStorage.getItem('token');
        // Additional validation or expiry check can be performed here
        if (token) {
            setAuthState(prevState => ({
                ...prevState,
                isLoggedIn: true, // Assume token is still valid
                token: token,
                username: localStorage.getItem('username'),
                usertype: localStorage.getItem('usertype'),
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

            const text = await response.text(); // First, get the text
            try {
                const data = JSON.parse(text); // Try parsing it
                if (response.ok) {
                    localStorage.setItem('token', data.token);
                    localStorage.setItem('username', data.username);
                    localStorage.setItem('usertype', data.usertype);
                    setAuthState({
                        isLoggedIn: true,
                        token: data.token,
                        username: data.username,
                        usertype: data.usertype,
                    });
                    return true;
                } else {
                    throw new Error(data.message || 'Login failed');
                }
            } catch (e) {
                throw new Error('Server response is not valid JSON or password was incorrect: ' + text);
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
        });
    };

    return (
        <AuthContext.Provider value={{ ...authState, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export { useAuth, AuthProvider, AuthContext };
