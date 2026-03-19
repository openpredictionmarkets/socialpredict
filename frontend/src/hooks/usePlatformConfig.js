import { useState, useEffect } from 'react';
import { API_URL } from '../config';

/**
 * Fetches /v0/setup/frontend once on mount.
 * Returns { platformPrivate, loading } — safe to gate routes behind.
 */
export function usePlatformConfig() {
  const [platformPrivate, setPlatformPrivate] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    fetch(`${API_URL}/v0/setup/frontend`)
      .then((r) => r.json())
      .then((data) => {
        if (!cancelled) {
          setPlatformPrivate(!!data.platformPrivate);
        }
      })
      .catch(() => {
        // On error, default to public (fail open)
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return { platformPrivate, loading };
}

export default usePlatformConfig;
