import { useState, useEffect } from 'react';
import { apiRequest } from '../api/httpClient';
import { useAuth } from '../helpers/AuthContent';

const useUserData = (username, usePrivateProfile = false) => {
  const [userData, setUserData] = useState(null);
  const [userLoading, setUserLoading] = useState(true);
  const [userError, setUserError] = useState(null);
  const { token } = useAuth();

  useEffect(() => {
    let ignore = false;

    const fetchUserData = async () => {
      try {
        const path = usePrivateProfile ? '/v0/privateprofile' : `/v0/userinfo/${username}`;
        const data = await apiRequest(path, {
          authenticated: true,
          authToken: token,
          fallbackMessage: 'Failed to fetch user data',
        });
        if (!ignore) {
          setUserData(data);
        }
      } catch (error) {
        console.error('Error fetching user data:', error);
        if (!ignore) {
          setUserError(error.toString());
        }
      } finally {
        if (!ignore) {
          setUserLoading(false);
        }
      }
    };

    if (!token) {
      setUserLoading(false);
      return () => {
        ignore = true;
      };
    }

    if (username || usePrivateProfile) {
      fetchUserData();
    }

    return () => {
      ignore = true;
    };
  }, [username, usePrivateProfile, token]);

  return { userData, userLoading, userError };
};

export default useUserData;
