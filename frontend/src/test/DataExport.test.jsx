import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import DataExport from '../components/layouts/admin/DataExport';

vi.mock('../config', () => ({ API_URL: 'http://localhost:8080' }));

describe('DataExport', () => {
  beforeEach(() => {
    localStorage.setItem('token', 'test-token');
    global.URL.createObjectURL = vi.fn(() => 'blob:mock');
    global.URL.revokeObjectURL = vi.fn();
  });

  afterEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('renders three download buttons', () => {
    render(<DataExport />);
    const buttons = screen.getAllByText('Download CSV');
    expect(buttons).toHaveLength(3);
  });

  it('renders labels for bets, markets, and users', () => {
    render(<DataExport />);
    expect(screen.getByText('Bets')).toBeInTheDocument();
    expect(screen.getByText('Markets')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
  });

  it('shows Downloading… while request is in flight', async () => {
    let resolveFetch;
    global.fetch = vi.fn(
      () =>
        new Promise((resolve) => {
          resolveFetch = resolve;
        })
    );

    render(<DataExport />);
    const [betsBtn] = screen.getAllByText('Download CSV');
    fireEvent.click(betsBtn);

    await waitFor(() =>
      expect(screen.getByText('Downloading…')).toBeInTheDocument()
    );

    // Cleanup: resolve to avoid unhandled promise
    resolveFetch({ ok: true, headers: { get: () => '' }, blob: async () => new Blob() });
  });

  it('shows error message on fetch failure', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: false,
        status: 403,
        text: async () => 'Forbidden',
      })
    );

    render(<DataExport />);
    const [betsBtn] = screen.getAllByText('Download CSV');
    fireEvent.click(betsBtn);

    await waitFor(() =>
      expect(screen.getByText('Forbidden')).toBeInTheDocument()
    );
  });

  it('triggers download link on success', async () => {
    const mockBlob = new Blob(['id,username\n1,alice'], { type: 'text/csv' });
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        headers: { get: (h) => (h === 'Content-Disposition' ? 'attachment; filename="bets_test.csv"' : '') },
        blob: async () => mockBlob,
      })
    );

    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});

    render(<DataExport />);
    const [betsBtn] = screen.getAllByText('Download CSV');
    fireEvent.click(betsBtn);

    await waitFor(() => expect(clickSpy).toHaveBeenCalled());
    expect(URL.createObjectURL).toHaveBeenCalledWith(mockBlob);
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:mock');
  });
});
