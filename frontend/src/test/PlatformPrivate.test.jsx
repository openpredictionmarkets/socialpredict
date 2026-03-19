import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';

vi.mock('../config', () => ({ API_URL: 'http://localhost:8080' }));

// Stub all page components so tests are fast
vi.mock('../pages/markets/Markets', () => ({ default: () => <div>Markets page</div> }));
vi.mock('../pages/polls/Polls', () => ({ default: () => <div>Polls page</div> }));
vi.mock('../pages/about/About', () => ({ default: () => <div>About page</div> }));
vi.mock('../pages/stats/Stats', () => ({ default: () => <div>Stats page</div> }));
vi.mock('../pages/user/User', () => ({ default: () => <div>User page</div> }));
vi.mock('../pages/marketDetails/MarketDetails', () => ({ default: () => <div>MarketDetails page</div> }));
vi.mock('../pages/home/Home', () => ({ default: () => <div>Home page</div> }));
vi.mock('../pages/changepassword/ChangePassword', () => ({ default: () => <div>ChangePassword</div> }));
vi.mock('../pages/profile/Profile', () => ({ default: () => <div>Profile</div> }));
vi.mock('../pages/notifications/Notifications', () => ({ default: () => <div>Notifications</div> }));
vi.mock('../pages/create/Create', () => ({ default: () => <div>Create</div> }));
vi.mock('../pages/style/Style', () => ({ default: () => <div>Style</div> }));
vi.mock('../pages/admin/AdminDashboard', () => ({ default: () => <div>AdminDashboard</div> }));
vi.mock('../pages/notfound/NotFound', () => ({ default: () => <div>Not Found</div> }));

// We control auth and platform config in each test
let mockAuthValue = { username: null, usertype: null, changePasswordNeeded: false };
let mockPlatformConfig = { platformPrivate: false, loading: false };

vi.mock('../helpers/AuthContent', () => ({
  useAuth: () => mockAuthValue,
}));

vi.mock('../hooks/usePlatformConfig', () => ({
  usePlatformConfig: () => mockPlatformConfig,
}));

import AppRoutes from '../helpers/AppRoutes';

function renderAt(path) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AppRoutes />
    </MemoryRouter>
  );
}

describe('AppRoutes — platform private gating', () => {
  beforeEach(() => {
    mockAuthValue = { username: null, usertype: null, changePasswordNeeded: false };
    mockPlatformConfig = { platformPrivate: false, loading: false };
  });

  it('shows Markets page when platformPrivate=false and user is not logged in', () => {
    mockPlatformConfig = { platformPrivate: false, loading: false };
    renderAt('/markets');
    expect(screen.getByText('Markets page')).toBeInTheDocument();
  });

  it('redirects /markets to / when platformPrivate=true and user is not logged in', () => {
    mockPlatformConfig = { platformPrivate: true, loading: false };
    renderAt('/markets');
    expect(screen.getByText('Home page')).toBeInTheDocument();
    expect(screen.queryByText('Markets page')).not.toBeInTheDocument();
  });

  it('redirects /polls to / when platformPrivate=true and user is not logged in', () => {
    mockPlatformConfig = { platformPrivate: true, loading: false };
    renderAt('/polls');
    expect(screen.getByText('Home page')).toBeInTheDocument();
    expect(screen.queryByText('Polls page')).not.toBeInTheDocument();
  });

  it('redirects /about to / when platformPrivate=true and user is not logged in', () => {
    mockPlatformConfig = { platformPrivate: true, loading: false };
    renderAt('/about');
    expect(screen.getByText('Home page')).toBeInTheDocument();
    expect(screen.queryByText('About page')).not.toBeInTheDocument();
  });

  it('redirects /stats to / when platformPrivate=true and user is not logged in', () => {
    mockPlatformConfig = { platformPrivate: true, loading: false };
    renderAt('/stats');
    expect(screen.getByText('Home page')).toBeInTheDocument();
    expect(screen.queryByText('Stats page')).not.toBeInTheDocument();
  });

  it('shows Markets page when platformPrivate=true and user IS logged in', () => {
    mockAuthValue = { username: 'alice', usertype: 'USER', changePasswordNeeded: false };
    mockPlatformConfig = { platformPrivate: true, loading: false };
    renderAt('/markets');
    expect(screen.getByText('Markets page')).toBeInTheDocument();
  });
});
