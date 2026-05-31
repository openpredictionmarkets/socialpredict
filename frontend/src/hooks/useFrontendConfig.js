import { useEffect, useState } from 'react';
import { apiRequest } from '../api/httpClient';

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
        const data = await apiRequest('/v0/setup/frontend', {
          fallbackMessage: 'Unable to load frontend config.',
        });
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
