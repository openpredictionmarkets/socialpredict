import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route } from 'react-router-dom';
import { vi } from 'vitest';

vi.mock('../config', () => ({ API_URL: 'http://localhost:8080' }));
vi.mock('../helpers/AuthContent', () => ({
  useAuth: () => ({ isLoggedIn: false, username: null }),
}));

// Stub sub-components so the test stays focused on the not-found path
vi.mock('../components/tabs/SiteTabs', () => ({ default: () => <div>Tabs</div> }));
vi.mock('../components/layouts/profile/public/UserInfoTabContent', () => ({ default: () => null }));
vi.mock('../components/layouts/profile/public/PortfolioTabContent', () => ({ default: () => null }));
vi.mock('../components/layouts/profile/public/UserFinancialStatementsLayout', () => ({ default: () => null }));
vi.mock('../components/layouts/profile/public/BetHistoryTabContent', () => ({ default: () => null }));

import User from '../pages/user/User';

function renderUser(username) {
  return render(
    <MemoryRouter initialEntries={[`/user/${username}`]}>
      <Route path="/user/:username">
        <User />
      </Route>
    </MemoryRouter>
  );
}

describe('User page — not found (#569)', () => {
  it('shows "User not found" when API returns 404', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 404, json: vi.fn() });

    renderUser('ghostuser');

    await waitFor(() =>
      expect(screen.getByText('User not found.')).toBeTruthy()
    );
  });

  it('shows loading state initially', () => {
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {})); // never resolves

    renderUser('slowuser');

    expect(screen.getByText('Loading user profile...')).toBeTruthy();
  });

  it('renders profile when API returns user data', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      status: 200,
      json: vi.fn().mockResolvedValue({
        username: 'alice',
        displayname: 'Alice',
        personalEmoji: '🙂',
        usertype: 'regular',
        accountBalance: 1000,
        initialAccountBalance: 1000,
      }),
    });

    renderUser('alice');

    await waitFor(() => expect(screen.getByText('Tabs')).toBeTruthy());
  });
});
