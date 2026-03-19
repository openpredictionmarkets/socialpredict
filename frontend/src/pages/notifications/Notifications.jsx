import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';

function Notifications() {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const token = localStorage.getItem('token');

  const fetchNotifications = useCallback(async () => {
    if (!token) {
      setLoading(false);
      return;
    }
    try {
      const response = await fetch(`${API_URL}/v0/notifications`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) throw new Error('Failed to fetch notifications');
      const data = await response.json();
      setNotifications(data || []);
    } catch (err) {
      setError('Could not load notifications.');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  const markAllRead = async () => {
    if (!token) return;
    try {
      await fetch(`${API_URL}/v0/notifications/read-all`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${token}` },
      });
      setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })));
    } catch {
      // non-fatal
    }
  };

  const unreadCount = notifications.filter((n) => !n.isRead).length;

  if (!token) {
    return (
      <div className='p-6 bg-primary-background shadow-md rounded-lg'>
        <h1 className='text-lg font-medium text-gray-300 mb-4'>Notifications</h1>
        <p className='text-gray-400'>Please log in to view your notifications.</p>
      </div>
    );
  }

  return (
    <div className='p-6 bg-primary-background shadow-md rounded-lg'>
      <div className='flex items-center justify-between mb-4'>
        <h1 className='text-lg font-medium text-gray-300'>
          Notifications
          {unreadCount > 0 && (
            <span className='ml-2 px-2 py-0.5 text-xs font-bold bg-blue-600 text-white rounded-full'>
              {unreadCount}
            </span>
          )}
        </h1>
        {unreadCount > 0 && (
          <button
            onClick={markAllRead}
            className='text-sm text-blue-400 hover:text-blue-300 transition-colors'
          >
            Mark all as read
          </button>
        )}
      </div>

      {error && <p className='text-red-400 text-sm mb-3'>{error}</p>}

      {loading ? (
        <p className='text-gray-400 text-sm'>Loading…</p>
      ) : notifications.length === 0 ? (
        <p className='text-gray-400 text-sm'>You have no notifications.</p>
      ) : (
        <ul className='space-y-2'>
          {notifications.map((n) => (
            <li
              key={n.id}
              className={`rounded-lg p-3 border-l-4 transition-colors ${
                n.isRead
                  ? 'bg-gray-800 border-gray-600'
                  : 'bg-gray-700 border-blue-500'
              }`}
            >
              <p className='text-gray-200 text-sm'>{n.message}</p>
              <div className='flex items-center justify-between mt-1'>
                <p className='text-gray-500 text-xs'>
                  {new Date(n.createdAt).toLocaleString()}
                </p>
                {n.marketId > 0 && (
                  <Link
                    to={`/market/${n.marketId}`}
                    className='text-blue-400 hover:text-blue-300 text-xs transition-colors'
                  >
                    View market →
                  </Link>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default Notifications;
