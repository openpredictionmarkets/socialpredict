import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';
import {
    getApiErrorMessage,
    parseApiResponseText,
    unwrapApiResponse,
} from '../utils/apiResponse';

const decodeJwtPayload = (token) => {
    if (!token) return null;

    try {
        const [, payload] = token.split('.');
        if (!payload) return null;

        const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
        const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), '=');
        return JSON.parse(window.atob(padded));
    } catch (error) {
        console.error('Failed to decode JWT payload:', error);
        return null;
    }
};

const getStoredAuthState = () => {
    const token = localStorage.getItem('token');
    const tokenPayload = decodeJwtPayload(token);
    const storedUsername = localStorage.getItem('username');
    const normalizedUsername = storedUsername && storedUsername !== 'undefined'
        ? storedUsername
        : tokenPayload?.username || null;

    return {
        isLoggedIn: Boolean(token),
        token,
        username: normalizedUsername,
        usertype: localStorage.getItem('usertype'),
        changePasswordNeeded: false,
    };
};

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
    const [authState, setAuthState] = useState(getStoredAuthState);

    const loginReasonMessages = {
        AUTHORIZATION_DENIED: 'Invalid username or password.',
        VALIDATION_FAILED: 'Username must use lowercase letters and numbers, and password cannot be empty.',
        INVALID_REQUEST: 'Invalid login request.',
        INVALID_TOKEN: 'Your session is invalid. Please try again.',
        PASSWORD_CHANGE_REQUIRED: 'Password change required before continuing.',
    };

    useEffect(() => {
        if (authState.isLoggedIn && authState.usertype) {
            // Redirect or perform other actions based on usertype
        }
    }, [authState.isLoggedIn, authState.usertype]);

    useEffect(() => {
        localStorage.removeItem('changePasswordNeeded');
        const storedState = getStoredAuthState();
        if (storedState.token) {
            if (storedState.username) {
                localStorage.setItem('username', storedState.username);
            }
            setAuthState(storedState);
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
            const data = parseApiResponseText(text);

            if (response.ok) {
                const authData = unwrapApiResponse(data);

                localStorage.setItem('token', authData.token);
                localStorage.setItem('username', authData.username);
                localStorage.setItem('usertype', authData.usertype);
                setAuthState({
                    isLoggedIn: true,
                    token: authData.token,
                    username: authData.username,
                    usertype: authData.usertype,
                    changePasswordNeeded: authData.mustChangePassword,
                });
                return { success: true, mustChangePassword: authData.mustChangePassword };
            } else {
                const errorMessage = getApiErrorMessage(
                    response,
                    data,
                    `Login failed with status ${response.status}.`,
                    loginReasonMessages,
                );
                throw new Error(errorMessage);
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    };


    const logout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        localStorage.removeItem('usertype');
        localStorage.removeItem('changePasswordNeeded');
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
