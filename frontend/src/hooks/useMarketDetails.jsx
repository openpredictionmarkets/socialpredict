import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom/cjs/react-router-dom';
import { API_URL } from '../config';

const calculateCurrentProbability = (details) => {
  if (!details || !details.probabilityChanges) return 0;

  const currentProbability =
    details.probabilityChanges.length > 0
      ? details.probabilityChanges[details.probabilityChanges.length - 1]
          .probability
      : details.market.initialProbability;

  return parseFloat(currentProbability.toFixed(3));
};

export const useMarketDetails = () => {
  const [details, setDetails] = useState(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState(null);
  const [currentProbability, setCurrentProbability] = useState(0);
  const { marketId } = useParams();
  const [triggerRefresh, setTriggerRefresh] = useState(false);

  useEffect(() => {
    const fetchedToken = localStorage.getItem('token');
    setToken(fetchedToken);
    setIsLoggedIn(!!fetchedToken);
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(`${API_URL}/api/v0/markets/${marketId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch market data');
        }
        const data = await response.json();
        setDetails(data);
        setCurrentProbability(calculateCurrentProbability(data));
      } catch (error) {
        console.error('Error fetching market data:', error);
      }
    };

    fetchData();
  }, [marketId, triggerRefresh]);

  const refetchData = () => {
    setTriggerRefresh((prev) => !prev);
  };

  return { details, isLoggedIn, token, refetchData, currentProbability };
};
