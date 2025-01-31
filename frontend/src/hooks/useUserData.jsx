import { useState, useEffect } from 'react';
import { API_URL } from '../config.js';

const useUserData = (username) => {
  const [userData, setUserData] = useState(null);
  const [userLoading, setUserLoading] = useState(true);
  const [userError, setUserError] = useState(null);

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const response = await fetch(`${API_URL}/api/v0/userinfo/${username}`);
        if (!response.ok) {
          throw new Error('Failed to fetch user data');
        }
        const data = await response.json();
        setUserData(data);
      } catch (error) {
        console.error('Error fetching user data:', error);
        setUserError(error.toString());
      } finally {
        setUserLoading(false);
      }
    };

    if (username) {
      fetchUserData();
    }
  }, [username]);

  return { userData, userLoading, userError };
};

export default useUserData;
