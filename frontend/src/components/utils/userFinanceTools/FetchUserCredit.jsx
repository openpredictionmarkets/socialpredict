import React, { useState, useEffect } from 'react';
import { apiRequest } from '../../../api/httpClient';
import { useAuth } from '../../../helpers/AuthContent';

export const USER_CREDIT_REFRESH_EVENT = 'user-credit-refresh';

const useUserCredit = (username) => {
    const [userCredit, setUserCredit] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { token } = useAuth();

    useEffect(() => {
      const fetchCredit = async () => {
        setLoading(true);
        setError(null);

        try {
          const data = await apiRequest(`/v0/usercredit/${username}`, {
            authenticated: true,
            authToken: token,
            fallbackMessage: 'Failed to fetch user credit',
          });
          setUserCredit(data.credit);
        } catch (error) {
          console.error('Error fetching user credit:', error);
          setError(error.toString());
        } finally {
          setLoading(false);
        }
      };

      const handleCreditRefresh = () => {
        if (username && token) {
          fetchCredit();
        }
      };

      window.addEventListener(USER_CREDIT_REFRESH_EVENT, handleCreditRefresh);

      if (username && token) {
        fetchCredit();
      } else if (!username || !token) {
        setLoading(false);
      }

      return () => {
        window.removeEventListener(USER_CREDIT_REFRESH_EVENT, handleCreditRefresh);
      };
    }, [username, token]);

    return { userCredit, loading, error };
  };

  export default useUserCredit;
