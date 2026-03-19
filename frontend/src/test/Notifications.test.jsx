import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';

// Stub react-router-dom Link
vi.mock('react-router-dom', () => ({
  Link: ({ children, to }) => <a href={to}>{children}</a>,
}));

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
  localStorage.setItem('token', 'fake-token');
});

import Notifications from '../pages/notifications/Notifications';

const emptyNotifs = () =>
  Promise.resolve({ ok: true, json: () => Promise.resolve([]) });

const notifsResponse = (items) =>
  Promise.resolve({ ok: true, json: () => Promise.resolve(items) });

describe('Notifications page', () => {
  it('shows login prompt when no token', async () => {
    localStorage.removeItem('token');
    render(<Notifications />);
    expect(screen.getByText(/please log in/i)).toBeTruthy();
  });

  it('shows "no notifications" for empty list', async () => {
    mockFetch.mockReturnValueOnce(emptyNotifs());
    render(<Notifications />);
    await waitFor(() =>
      expect(screen.getByText(/you have no notifications/i)).toBeTruthy()
    );
  });

  it('renders notification messages', async () => {
    mockFetch.mockReturnValueOnce(
      notifsResponse([
        {
          id: 1,
          type: 'market_resolved',
          marketId: 5,
          message: 'Market "Will it rain?" resolved: YES',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<Notifications />);
    await waitFor(() =>
      expect(
        screen.getByText('Market "Will it rain?" resolved: YES')
      ).toBeTruthy()
    );
  });

  it('shows unread count badge when unread notifications exist', async () => {
    mockFetch.mockReturnValueOnce(
      notifsResponse([
        {
          id: 1,
          type: 'market_resolved',
          marketId: 1,
          message: 'Resolved',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
        {
          id: 2,
          type: 'market_resolved',
          marketId: 2,
          message: 'Also resolved',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<Notifications />);
    await waitFor(() => screen.getByText('Resolved'));
    // Badge should show 2
    expect(screen.getByText('2')).toBeTruthy();
  });

  it('shows "Mark all as read" button when there are unread notifications', async () => {
    mockFetch.mockReturnValueOnce(
      notifsResponse([
        {
          id: 1,
          type: 'market_resolved',
          marketId: 1,
          message: 'Unread',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<Notifications />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark all as read/i })).toBeTruthy()
    );
  });

  it('does not show "Mark all as read" when all are read', async () => {
    mockFetch.mockReturnValueOnce(
      notifsResponse([
        {
          id: 1,
          type: 'market_resolved',
          marketId: 1,
          message: 'Already read',
          isRead: true,
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<Notifications />);
    await waitFor(() => screen.getByText('Already read'));
    expect(screen.queryByRole('button', { name: /mark all as read/i })).toBeNull();
  });
});
