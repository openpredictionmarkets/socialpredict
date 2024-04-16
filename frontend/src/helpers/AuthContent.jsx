import { API_URL } from './../config';
import React, { createContext, useContext, useState } from 'react';

const AuthContext = createContext(null);

function useAuth() {
    return useContext(AuthContext);
}

const AuthProvider = ({ children }) => {
    const [authState, setAuthState] = useState({
        isLoggedIn: false,
        token: localStorage.getItem('token'),
        username: localStorage.getItem('username'),
    });

    const login = async (username, password) => {
        console.log('login function received:', username, password);
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
                setAuthState({
                    isLoggedIn: true,
                    token: data.token,
                    username: username,
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
        localStorage.removeItem('token');
        localStorage.removeItem('username');
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
