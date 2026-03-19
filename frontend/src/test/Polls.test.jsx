import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import Polls from '../pages/polls/Polls';

vi.mock('../config', () => ({ API_URL: 'http://localhost:8080' }));

const mockAuth = { username: null };
vi.mock('../helpers/AuthContent', () => ({
  useAuth: () => mockAuth,
}));

const makePoll = (overrides = {}) => ({
  id: 1,
  creatorUsername: 'alice',
  question: 'Will it rain?',
  description: '',
  isClosed: false,
  yesCount: 3,
  noCount: 1,
  userVote: '',
  createdAt: new Date().toISOString(),
  ...overrides,
});

describe('Polls page', () => {
  beforeEach(() => {
    mockAuth.username = null;
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('shows loading state initially', () => {
    global.fetch = vi.fn(() => new Promise(() => {}));
    render(<Polls />);
    expect(screen.getByText(/loading polls/i)).toBeInTheDocument();
  });

  it('shows empty message when no polls', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [] })
    );
    render(<Polls />);
    await waitFor(() => expect(screen.getByText(/no open polls/i)).toBeInTheDocument());
  });

  it('renders poll question and vote counts', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [makePoll()] })
    );
    render(<Polls />);
    await waitFor(() => expect(screen.getByText('Will it rain?')).toBeInTheDocument());
    expect(screen.getByText(/4 vote/i)).toBeInTheDocument();
  });

  it('does not show vote buttons when not logged in', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [makePoll()] })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText('Will it rain?'));
    expect(screen.queryByText(/vote yes/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/vote no/i)).not.toBeInTheDocument();
    expect(screen.getByText(/log in to vote/i)).toBeInTheDocument();
  });

  it('shows vote buttons when logged in and not voted', async () => {
    mockAuth.username = 'alice';
    localStorage.setItem('token', 'tok');
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [makePoll()] })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText('Will it rain?'));
    expect(screen.getByText(/vote yes/i)).toBeInTheDocument();
    expect(screen.getByText(/vote no/i)).toBeInTheDocument();
  });

  it('hides vote buttons when user already voted', async () => {
    mockAuth.username = 'alice';
    localStorage.setItem('token', 'tok');
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: async () => [makePoll({ userVote: 'YES' })],
      })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText('Will it rain?'));
    expect(screen.queryByText(/vote yes/i)).not.toBeInTheDocument();
    expect(screen.getByText(/You voted/i)).toBeInTheDocument();
  });

  it('shows Closed badge for closed polls', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: async () => [makePoll({ isClosed: true })],
      })
    );
    render(<Polls />);
    await waitFor(() => expect(screen.getByText('Closed')).toBeInTheDocument());
  });

  it('shows create poll form when logged in', async () => {
    mockAuth.username = 'alice';
    localStorage.setItem('token', 'tok');
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [] })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText(/no open polls/i));
    expect(screen.getByText(/create a poll/i)).toBeInTheDocument();
  });

  it('hides create poll form when not logged in', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [] })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText(/no open polls/i));
    expect(screen.queryByText(/create a poll/i)).not.toBeInTheDocument();
  });

  it('shows close poll button for creator', async () => {
    mockAuth.username = 'alice';
    localStorage.setItem('token', 'tok');
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => [makePoll()] })
    );
    render(<Polls />);
    await waitFor(() => screen.getByText('Will it rain?'));
    expect(screen.getByText(/close poll/i)).toBeInTheDocument();
  });
});
