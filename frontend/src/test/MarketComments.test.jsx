import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import MarketComments from '../components/comments/MarketComments';

// Stub global fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

const emptyResponse = () =>
  Promise.resolve({ ok: true, json: () => Promise.resolve([]) });

const commentsResponse = (items) =>
  Promise.resolve({ ok: true, json: () => Promise.resolve(items) });

beforeEach(() => {
  mockFetch.mockReset();
  localStorage.setItem('username', 'alice');
});

describe('MarketComments', () => {
  it('shows "No comments yet" when list is empty', async () => {
    mockFetch.mockReturnValueOnce(emptyResponse());
    render(<MarketComments marketId={1} isLoggedIn={false} token={null} />);
    await waitFor(() =>
      expect(screen.getByText(/no comments yet/i)).toBeTruthy()
    );
  });

  it('renders existing comments', async () => {
    mockFetch.mockReturnValueOnce(
      commentsResponse([
        {
          id: 1,
          marketId: 1,
          username: 'bob',
          content: 'First comment',
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<MarketComments marketId={1} isLoggedIn={false} token={null} />);
    await waitFor(() =>
      expect(screen.getByText('First comment')).toBeTruthy()
    );
    expect(screen.getByText('bob')).toBeTruthy();
  });

  it('shows comment form when logged in', async () => {
    mockFetch.mockReturnValueOnce(emptyResponse());
    render(<MarketComments marketId={1} isLoggedIn={true} token='tok' />);
    await waitFor(() =>
      expect(screen.getByPlaceholderText(/add a comment/i)).toBeTruthy()
    );
  });

  it('hides comment form when logged out', async () => {
    mockFetch.mockReturnValueOnce(emptyResponse());
    render(<MarketComments marketId={1} isLoggedIn={false} token={null} />);
    await waitFor(() => screen.getByText(/no comments yet/i));
    expect(screen.queryByPlaceholderText(/add a comment/i)).toBeNull();
  });

  it('submit button is disabled when textarea is empty', async () => {
    mockFetch.mockReturnValueOnce(emptyResponse());
    render(<MarketComments marketId={1} isLoggedIn={true} token='tok' />);
    await waitFor(() => screen.getByPlaceholderText(/add a comment/i));
    const button = screen.getByRole('button', { name: /post comment/i });
    expect(button.disabled).toBe(true);
  });

  it('enables submit button when text is entered', async () => {
    mockFetch.mockReturnValueOnce(emptyResponse());
    render(<MarketComments marketId={1} isLoggedIn={true} token='tok' />);
    await waitFor(() => screen.getByPlaceholderText(/add a comment/i));
    const textarea = screen.getByPlaceholderText(/add a comment/i);
    fireEvent.change(textarea, { target: { value: 'Hello world' } });
    const button = screen.getByRole('button', { name: /post comment/i });
    expect(button.disabled).toBe(false);
  });

  it('shows delete button only for own comments when logged in', async () => {
    mockFetch.mockReturnValueOnce(
      commentsResponse([
        {
          id: 1,
          marketId: 1,
          username: 'alice', // same as localStorage username
          content: 'My comment',
          createdAt: new Date().toISOString(),
        },
        {
          id: 2,
          marketId: 1,
          username: 'bob',
          content: "Bob's comment",
          createdAt: new Date().toISOString(),
        },
      ])
    );
    render(<MarketComments marketId={1} isLoggedIn={true} token='tok' />);
    await waitFor(() => screen.getByText('My comment'));

    // Only one delete button (for alice's comment, not bob's)
    const deleteButtons = screen.queryAllByRole('button', {
      name: /delete comment/i,
    });
    expect(deleteButtons.length).toBe(1);
  });
});
