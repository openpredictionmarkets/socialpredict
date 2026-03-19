import { useState, useEffect } from 'react';
import { API_URL } from '../config';

/**
 * Polls the unread notification count every 60 seconds for the logged-in user.
 * Returns { unreadCount } — safe to render in the sidebar badge.
 */
export function useUnreadNotifications(isLoggedIn) {
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    if (!isLoggedIn) {
      setUnreadCount(0);
      return;
    }

    const token = localStorage.getItem('token');
    if (!token) return;

    const fetch_count = async () => {
      try {
        const response = await fetch(`${API_URL}/v0/notifications/unread`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (response.ok) {
          const data = await response.json();
          setUnreadCount(data.unreadCount ?? 0);
        }
      } catch {
        // silent — badge just stays at previous value
      }
    };

    fetch_count();
    const interval = setInterval(fetch_count, 60_000);
    return () => clearInterval(interval);
  }, [isLoggedIn]);

  return { unreadCount };
}
