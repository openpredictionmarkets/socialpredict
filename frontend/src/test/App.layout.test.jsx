import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

// Stub heavy dependencies so this stays a unit test
vi.mock('../helpers/AuthContent', () => ({
  AuthProvider: ({ children }) => <>{children}</>,
  useAuth: () => ({
    isLoggedIn: false,
    usertype: null,
    logout: vi.fn(),
    changePasswordNeeded: false,
    username: null,
  }),
}));

vi.mock('../helpers/AppRoutes', () => ({
  default: () => <div data-testid="routes" />,
}));

vi.mock('../components/sidebar/Sidebar', () => ({
  default: () => <nav data-testid="sidebar" />,
}));

vi.mock('../components/footer/Footer', () => ({
  default: () => <footer data-testid="footer" />,
}));

import App from '../App';

describe('App layout — issue #9 toolbar overlap fix', () => {
  it('main content has pb-16 so mobile bottom nav does not overlap page content', () => {
    const { container } = render(<App />);
    const main = container.querySelector('main');
    expect(main).not.toBeNull();
    expect(main.className).toContain('pb-16');
  });

  it('main content resets bottom padding on md+ screens (md:pb-6) so desktop is unaffected', () => {
    const { container } = render(<App />);
    const main = container.querySelector('main');
    expect(main.className).toContain('md:pb-6');
  });
});
