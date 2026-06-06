import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiRequest } from '../api/httpClient';
import { authStorage } from '../api/authStorage';

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
    const token = authStorage.getToken();
    const tokenPayload = decodeJwtPayload(token);
    const storedUsername = authStorage.getUsername();
    const normalizedUsername = storedUsername && storedUsername !== 'undefined'
        ? storedUsername
        : tokenPayload?.username || null;

    return {
        isLoggedIn: Boolean(token),
        token,
        username: normalizedUsername,
        usertype: authStorage.getUsertype(),
        moderatorStatus: authStorage.getModeratorStatus(),
        changePasswordNeeded: false,
    };
};

const AuthContext = createContext({
    username: null,
    setUsername: () => {},
    isLoggedIn: false,
    usertype: null,
    moderatorStatus: null,
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
        if (!authState.isLoggedIn || !authState.token || authState.changePasswordNeeded) {
            return undefined;
        }

        let ignore = false;
        const refreshProfile = async () => {
            try {
                const profile = await apiRequest('/v0/privateprofile', {
                    authenticated: true,
                    authToken: authState.token,
                    fallbackMessage: 'Failed to refresh user profile.',
                });
                if (ignore) return;

                authStorage.saveLogin({
                    token: authState.token,
                    username: profile.username || authState.username,
                    usertype: profile.usertype || authState.usertype,
                    moderatorStatus: profile.moderatorStatus || authState.moderatorStatus,
                });
                setAuthState((current) => ({
                    ...current,
                    username: profile.username || current.username,
                    usertype: profile.usertype || current.usertype,
                    moderatorStatus: profile.moderatorStatus || current.moderatorStatus,
                }));
            } catch (error) {
                console.error('Failed to refresh auth profile:', error);
            }
        };

        refreshProfile();
        return () => {
            ignore = true;
        };
    }, [authState.isLoggedIn, authState.token, authState.changePasswordNeeded]);

    useEffect(() => {
        authStorage.clearLegacyPasswordChangeFlag();
        const storedState = getStoredAuthState();
        if (storedState.token) {
            if (storedState.username) {
                authStorage.setUsername(storedState.username);
            }
            setAuthState(storedState);
        }
    }, []);

    const login = async (username, password) => {
        try {
            const authData = await apiRequest('/v0/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
                fallbackMessage: 'Login failed. Please try again.',
                reasonMessages: loginReasonMessages,
            });

            authStorage.saveLogin(authData);
            setAuthState({
                isLoggedIn: true,
                token: authData.token,
                username: authData.username,
                usertype: authData.usertype,
                moderatorStatus: authData.moderatorStatus,
                changePasswordNeeded: authData.mustChangePassword,
            });
            return { success: true, mustChangePassword: authData.mustChangePassword };
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    };


    const logout = () => {
        authStorage.clear();
        setAuthState({
            isLoggedIn: false,
            token: null,
            username: null,
            usertype: null,
            moderatorStatus: null,
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
