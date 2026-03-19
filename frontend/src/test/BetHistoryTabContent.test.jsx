import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';

vi.mock('react-router-dom', () => ({
  Link: ({ children, to }) => <a href={to}>{children}</a>,
}));

vi.mock('../components/loaders/LoadingSpinner', () => ({
  default: () => <div data-testid='spinner' />,
}));

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => mockFetch.mockReset());

import BetHistoryTabContent from '../components/layouts/profile/public/BetHistoryTabContent';

const betsResponse = (items) =>
  Promise.resolve({ ok: true, json: () => Promise.resolve(items) });

describe('BetHistoryTabContent', () => {
  it('shows "No bets placed yet" for empty list', async () => {
    mockFetch.mockReturnValueOnce(betsResponse([]));
    render(<BetHistoryTabContent username='alice' />);
    await waitFor(() =>
      expect(screen.getByText(/no bets placed yet/i)).toBeTruthy()
    );
  });

  it('renders bet rows with market title, action, outcome, amount', async () => {
    mockFetch.mockReturnValueOnce(
      betsResponse([
        {
          id: 1,
          marketId: 5,
          questionTitle: 'Will it rain?',
          action: 'BUY',
          outcome: 'YES',
          amount: 100,
          placedAt: new Date().toISOString(),
        },
      ])
    );
    render(<BetHistoryTabContent username='alice' />);
    await waitFor(() => screen.getByText('Will it rain?'));
    expect(screen.getByText('BUY')).toBeTruthy();
    expect(screen.getByText('YES')).toBeTruthy();
    expect(screen.getByText('100 🪙')).toBeTruthy();
  });

  it('shows absolute amount for SELL bets', async () => {
    mockFetch.mockReturnValueOnce(
      betsResponse([
        {
          id: 2,
          marketId: 5,
          questionTitle: 'Will it snow?',
          action: 'SELL',
          outcome: 'NO',
          amount: -50,
          placedAt: new Date().toISOString(),
        },
      ])
    );
    render(<BetHistoryTabContent username='alice' />);
    await waitFor(() => screen.getByText('Will it snow?'));
    // Amount should be shown as absolute value
    expect(screen.getByText('50 🪙')).toBeTruthy();
    expect(screen.getByText('SELL')).toBeTruthy();
  });
});
