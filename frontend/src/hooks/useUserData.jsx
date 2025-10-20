import { useState, useEffect } from 'react';
import { API_URL } from '../config';
import { useAuth } from '../helpers/AuthContent';

const useUserData = (username, usePrivateProfile = false) => {
  const [userData, setUserData] = useState(null);
  const [userLoading, setUserLoading] = useState(true);
  const [userError, setUserError] = useState(null);
  const { token } = useAuth();

  useEffect(() => {
    const fetchUserData = async () => {
      try {
<<<<<<< HEAD
        let url, headers = {};
        
        if (usePrivateProfile) {
          // Use private profile endpoint for authenticated user's own profile
          url = `${API_URL}/api/v0/privateprofile`;
          headers = {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          };
        } else {
          // Use public user endpoint for viewing other users' profiles
          url = `${API_URL}/api/v0/userinfo/${username}`;
          if (token) {
            headers = {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            };
          }
        }

        const response = await fetch(url, { headers });
=======
        const response = await fetch(`${API_URL}/v0/userinfo/${username}`);
>>>>>>> main
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

    if (username || usePrivateProfile) {
      fetchUserData();
    }
  }, [username, usePrivateProfile, token]);

  return { userData, userLoading, userError };
};

export default useUserData;
