import { useState, useEffect } from 'react';
import { API_URL } from '../config';
import { useAuth } from '../helpers/AuthContent';

const usePortfolio = (username) => {
  const [portfolio, setPortfolio] = useState({ portfolioItems: [] });
  const [portfolioLoading, setPortfolioLoading] = useState(true);
  const [portfolioError, setPortfolioError] = useState(null);
  const { token } = useAuth();

  useEffect(() => {
    const fetchPortfolio = async () => {
      try {
<<<<<<< HEAD
        const headers = {};
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
          headers['Content-Type'] = 'application/json';
        }

        const response = await fetch(`${API_URL}/api/v0/portfolio/${username}`, { headers });
=======
        const response = await fetch(`${API_URL}/v0/portfolio/${username}`);
>>>>>>> main
        if (!response.ok) {
          throw new Error('Failed to fetch portfolio');
        }
        const data = await response.json();
        setPortfolio({ ...data, portfolioItems: data.portfolioItems || [] });
      } catch (error) {
        console.error('Error fetching portfolio:', error);
        setPortfolioError(error.toString());
      } finally {
        setPortfolioLoading(false);
      }
    };

    if (username && token) {
      fetchPortfolio();
    }
  }, [username, token]);

  return { portfolio, portfolioLoading, portfolioError };
};

export default usePortfolio;
