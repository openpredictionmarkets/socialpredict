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
                changePasswordNeeded: localStorage.getItem('changePasswordNeeded') === 'true', // assume password change needed until shown not
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

            const data = await response.json();
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
                throw new Error(data.message || 'Login failed');
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
