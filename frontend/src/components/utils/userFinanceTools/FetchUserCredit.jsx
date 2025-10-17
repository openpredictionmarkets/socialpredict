import React, { useState, useEffect } from 'react';
import { API_URL } from '../../../config';

const useUserCredit = (username) => {
    const [userCredit, setUserCredit] = useState(null);
    const [loading, setLoading] = useState(true); // Make sure this state is named correctly
    const [error, setError] = useState(null);

    useEffect(() => {
      const fetchCredit = async () => {
        try {
          const response = await fetch(`${API_URL}/v0/usercredit/${username}`);
          if (!response.ok) {
            throw new Error('Failed to fetch user credit');
          }
          const data = await response.json();;
          setUserCredit(data.credit);
        } catch (error) {
          console.error('Error fetching user credit:', error);
          setError(error.toString());
        } finally {
          setLoading(false);
        }
      };

      if (username) {
        fetchCredit();
      }
    }, [username]);

    return { userCredit, loading, error };
  };

  export default useUserCredit;
