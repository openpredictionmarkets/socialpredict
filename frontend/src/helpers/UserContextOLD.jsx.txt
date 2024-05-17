import React, { createContext, useState, useEffect } from 'react';

const UserContext = createContext({
  username: null,
  setUsername: () => {},
  isLoggedIn: false
});

export const UserProvider = ({ children }) => {
  const [username, setUsername] = useState(localStorage.getItem('username'));
  const isLoggedIn = username !== null;

  useEffect(() => {
    if (username) {
      localStorage.setItem('username', username);
    } else {
      localStorage.removeItem('username');
    }
  }, [username]);

  const contextValue = {
    username,
    setUsername,
    isLoggedIn
  };

  return (
    <UserContext.Provider value={contextValue}>
      {children}
    </UserContext.Provider>
  );
};

export default UserContext;
