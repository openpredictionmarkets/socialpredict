import { useEffect, useState } from 'react';
import { API_URL } from '../config';

const defaultFrontendConfig = {
  charts: { sigFigs: 2 },
  game: { mode: 'moderator' },
};

const useFrontendConfig = () => {
  const [frontendConfig, setFrontendConfig] = useState(defaultFrontendConfig);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadConfig = async () => {
      setLoading(true);
      setError('');
      try {
        const response = await fetch(`${API_URL}/v0/setup/frontend`);
        if (!response.ok) {
          throw new Error(`Frontend config failed with status ${response.status}`);
        }
        const data = await response.json();
        if (!ignore) {
          setFrontendConfig({
            ...defaultFrontendConfig,
            ...data,
            charts: { ...defaultFrontendConfig.charts, ...(data.charts || {}) },
            game: { ...defaultFrontendConfig.game, ...(data.game || {}) },
          });
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load frontend config.');
          setFrontendConfig(defaultFrontendConfig);
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadConfig();
    return () => {
      ignore = true;
    };
  }, []);

  return { frontendConfig, loading, error };
};

export default useFrontendConfig;
