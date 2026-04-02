import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import ActivityTabs from '../components/tabs/ActivityTabs';

vi.mock('../config', () => ({ API_URL: 'http://localhost:8080' }));

// Stub heavy child components to focus on bets lazy-load behaviour
vi.mock('../components/layouts/activity/positions/PositionsActivity', () => ({
  default: () => <div>Positions</div>,
}));
vi.mock('../components/layouts/activity/leaderboard/LeaderboardActivity', () => ({
  default: () => <div>Leaderboard</div>,
}));
vi.mock('../components/comments/MarketComments', () => ({
  default: () => <div>Comments</div>,
}));

const market = { id: 1, isResolved: false };

describe('ActivityTabs — lazy bets loading', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('does NOT call the bets API on initial render (Positions tab is default)', () => {
    global.fetch = vi.fn();
    render(
      <MemoryRouter>
        <ActivityTabs marketId={1} market={market} refreshTrigger={0} />
      </MemoryRouter>
    );
    // Fetch should not have been called yet — bets tab not open
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('calls the bets API when the Bets tab is clicked', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [] })
    );
    render(
      <MemoryRouter>
        <ActivityTabs marketId={1} market={market} refreshTrigger={0} />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByText('Bets'));

    await waitFor(() =>
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/v0/markets/bets/1')
      )
    );
  });

  it('does NOT re-fetch when switching back to Bets tab without refreshTrigger change', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [] })
    );

    render(
      <MemoryRouter>
        <ActivityTabs marketId={1} market={market} refreshTrigger={0} />
      </MemoryRouter>
    );

    // First open: should fetch
    fireEvent.click(screen.getByText('Bets'));
    await waitFor(() => expect(global.fetch).toHaveBeenCalledTimes(1));

    // Switch away
    fireEvent.click(screen.getByText('Positions'));
    // Switch back — should NOT re-fetch
    fireEvent.click(screen.getByText('Bets'));

    await new Promise((r) => setTimeout(r, 50));
    expect(global.fetch).toHaveBeenCalledTimes(1);
  });

  it('shows placeholder text before Bets tab is opened', () => {
    global.fetch = vi.fn(() => new Promise(() => {}));
    render(
      <MemoryRouter>
        <ActivityTabs marketId={1} market={market} refreshTrigger={0} />
      </MemoryRouter>
    );
    // Switch to Bets tab — shows placeholder until fetch resolves
    fireEvent.click(screen.getByText('Bets'));
    expect(screen.getByText(/loading bets/i)).toBeInTheDocument();
  });

  it('shows bets rows after fetch resolves', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: async () => [
          {
            username: 'alice',
            outcome: 'YES',
            amount: 50,
            probability: 0.6,
            placedAt: new Date().toISOString(),
          },
        ],
      })
    );

    render(
      <MemoryRouter>
        <ActivityTabs marketId={1} market={market} refreshTrigger={0} />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByText('Bets'));
    await waitFor(() => expect(screen.getByText('alice')).toBeInTheDocument());
  });
});
