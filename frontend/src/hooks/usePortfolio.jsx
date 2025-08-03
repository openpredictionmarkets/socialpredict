import { useState, useEffect } from 'react';
import { API_URL } from '../config';

const usePortfolio = (username) => {
  const [portfolio, setPortfolio] = useState({ portfolioItems: [] });
  const [portfolioLoading, setPortfolioLoading] = useState(true);
  const [portfolioError, setPortfolioError] = useState(null);

  useEffect(() => {
    const fetchPortfolio = async () => {
      try {
        const response = await fetch(`${API_URL}/v0/portfolio/${username}`);
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

    if (username) {
      fetchPortfolio();
    }
  }, [username]);

  return { portfolio, portfolioLoading, portfolioError };
};

export default usePortfolio;
