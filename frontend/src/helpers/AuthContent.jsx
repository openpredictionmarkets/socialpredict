import { API_URL } from './../config';
import React, { createContext, useContext, useState, useEffect } from 'react';

const unwrapApiResponse = (payload) => {
    if (payload && typeof payload === 'object' && 'ok' in payload) {
        if (payload.ok === false) {
            throw new Error(payload.reason || 'Request failed');
        }

        if (payload.ok === true && 'result' in payload) {
            return payload.result;
        }
    }

    return payload;
};

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
        changePasswordNeeded: localStorage.getItem('changePasswordNeeded') === 'true',
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

    useEffect(() => {
        if (authState.isLoggedIn && authState.usertype) {
            // Redirect or perform other actions based on usertype
        }
    }, [authState.isLoggedIn, authState.usertype]);

    useEffect(() => {
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
            let data = {};

            // Safely attempt to parse JSON
            try {
                data = JSON.parse(text);
            } catch (parseError) {
                // If JSON parsing fails, create a basic error object
                data = { error: text || 'Unknown error occurred' };
            }

            if (response.ok) {
                const authData = unwrapApiResponse(data);

                localStorage.setItem('token', authData.token);
                localStorage.setItem('username', authData.username);
                localStorage.setItem('usertype', authData.usertype);
                localStorage.setItem('changePasswordNeeded', authData.mustChangePassword);
                setAuthState({
                    isLoggedIn: true,
                    token: authData.token,
                    username: authData.username,
                    usertype: authData.usertype,
                    changePasswordNeeded: authData.mustChangePassword,
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
