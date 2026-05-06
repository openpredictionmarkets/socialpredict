import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';
import {
    getApiErrorMessage,
    parseApiResponseText,
    unwrapApiResponse,
} from '../utils/apiResponse';

const AUTH_STORAGE_KEYS = ['token', 'username', 'usertype', 'changePasswordNeeded'];

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

const clearStoredAuthState = () => {
    AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key));
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
        changePasswordNeeded: null,
        isAuthReady: !token,
    };
};

const AuthContext = createContext({
    username: null,
    setUsername: () => {},
    isLoggedIn: false,
    usertype: null,
    changePasswordNeeded: null,
    isAuthReady: false,
    login: async () => null,
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
        if (!authState.token || authState.isAuthReady) {
            return;
        }

        let isCancelled = false;

        // Restore password-change state from the API instead of persisting it in browser storage.
        const hydrateAuthState = async () => {
            try {
                const response = await fetch(`${API_URL}/v0/privateprofile`, {
                    headers: {
                        'Authorization': `Bearer ${authState.token}`,
                        'Content-Type': 'application/json',
                    },
                });
                const payload = parseApiResponseText(await response.text());

                if (response.ok) {
                    const profile = unwrapApiResponse(payload);

                    if (isCancelled) {
                        return;
                    }

                    localStorage.setItem('username', profile.username);
                    localStorage.setItem('usertype', profile.usertype);
                    localStorage.removeItem('changePasswordNeeded');

                    setAuthState((currentState) => {
                        if (currentState.token !== authState.token) {
                            return currentState;
                        }

                        return {
                            isLoggedIn: true,
                            token: authState.token,
                            username: profile.username,
                            usertype: profile.usertype,
                            changePasswordNeeded: profile.mustChangePassword,
                            isAuthReady: true,
                        };
                    });
                    return;
                }

                if (response.status === 403 && payload?.reason === 'PASSWORD_CHANGE_REQUIRED') {
                    if (isCancelled) {
                        return;
                    }

                    localStorage.removeItem('changePasswordNeeded');

                    setAuthState((currentState) => {
                        if (currentState.token !== authState.token) {
                            return currentState;
                        }

                        return {
                            ...currentState,
                            isLoggedIn: true,
                            changePasswordNeeded: true,
                            isAuthReady: true,
                        };
                    });
                    return;
                }

                throw new Error(getApiErrorMessage(
                    response,
                    payload,
                    'Failed to restore your session.',
                ));
            } catch (error) {
                console.error('Failed to restore auth state:', error);

                if (isCancelled) {
                    return;
                }

                clearStoredAuthState();
                setAuthState({
                    isLoggedIn: false,
                    token: null,
                    username: null,
                    usertype: null,
                    changePasswordNeeded: null,
                    isAuthReady: true,
                });
            }
        };

        hydrateAuthState();

        return () => {
            isCancelled = true;
        };
    }, [authState.isAuthReady, authState.token]);

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
                localStorage.removeItem('changePasswordNeeded');
                setAuthState({
                    isLoggedIn: true,
                    token: authData.token,
                    username: authData.username,
                    usertype: authData.usertype,
                    changePasswordNeeded: authData.mustChangePassword,
                    isAuthReady: true,
                });
                return authData;
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
        clearStoredAuthState();
        setAuthState({
            isLoggedIn: false,
            token: null,
            username: null,
            usertype: null,
            changePasswordNeeded: null,
            isAuthReady: true,
        });
    };

    return (
        <AuthContext.Provider value={{ ...authState, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export { useAuth, AuthProvider, AuthContext };
