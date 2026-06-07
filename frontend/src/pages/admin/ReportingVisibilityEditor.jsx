import React, { useEffect, useState } from 'react';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';

const adminMutationReasonMessages = {
  PASSWORD_CHANGE_REQUIRED: 'Your admin password must be changed before updating reporting visibility settings.',
  AUTHORIZATION_DENIED: 'Only admin users can update reporting visibility settings.',
  INVALID_TOKEN: 'Your session expired. Please log in again.',
};

const normalizeSettings = (data = {}) => ({
  systemMetricsPublic: data.systemMetricsPublic !== false,
  globalLeaderboardPublic: data.globalLeaderboardPublic !== false,
  updatedAt: data.updatedAt || '',
  version: data.version || 0,
});

function ToggleCard({ title, description, checked, onChange }) {
  return (
    <label className="block rounded-xl border border-gray-700 bg-gray-800 p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-white">{title}</h2>
          <p className="mt-2 text-sm leading-6 text-gray-300">{description}</p>
        </div>
        <input
          type="checkbox"
          checked={checked}
          onChange={(event) => onChange(event.target.checked)}
          className="mt-1 h-5 w-5 rounded border-gray-500 bg-gray-700 text-sky-400 focus:ring-sky-400"
        />
      </div>
      <p className="mt-4 text-xs uppercase tracking-[0.18em] text-gray-400">
        {checked ? 'Visible to logged-out visitors' : 'Requires login'}
      </p>
    </label>
  );
}

function ReportingVisibilityEditor() {
  const [settings, setSettings] = useState(normalizeSettings());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const data = await apiRequest('/v0/content/reporting-visibility', {
        fallbackMessage: 'Failed to load reporting visibility settings',
      });
      setSettings(normalizeSettings(data));
    } catch (err) {
      setError(err.message || 'Error loading reporting visibility settings');
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = (field, value) => {
    setSettings(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage('');
    setError('');

    try {
      const data = await authenticatedApiRequest('/v0/admin/content/reporting-visibility', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings),
        reasonMessages: adminMutationReasonMessages,
        fallbackMessage: 'Failed to save reporting visibility settings',
      });
      setSettings(normalizeSettings(data));
      setMessage('Reporting visibility settings saved successfully.');
    } catch (err) {
      setError(err.message || 'Error saving reporting visibility settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-primary-background p-8">
        <div className="mx-auto max-w-4xl">
          <p className="text-white">Loading reporting visibility settings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-primary-background p-8">
      <div className="mx-auto max-w-4xl space-y-6">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-sky-300">CMS</p>
          <h1 className="mt-2 text-3xl font-bold text-white">Reporting Visibility</h1>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-gray-300">
            Choose whether aggregate reporting pages are visible to logged-out visitors. User financial summaries remain login-only game transparency data.
          </p>
        </div>

        {message && (
          <div className="rounded bg-green-600 p-4 text-white">
            {message}
          </div>
        )}

        {error && (
          <div className="rounded bg-red-600 p-4 text-white">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <ToggleCard
            title="System Stats Public"
            description="Allows logged-out visitors to view aggregate system metrics. Turn this off when stats should require a logged-in game account."
            checked={settings.systemMetricsPublic}
            onChange={(value) => handleToggle('systemMetricsPublic', value)}
          />
          <ToggleCard
            title="Global Leaderboard Public"
            description="Allows logged-out visitors to view the global leaderboard. Turn this off when leaderboard transparency should be limited to logged-in users."
            checked={settings.globalLeaderboardPublic}
            onChange={(value) => handleToggle('globalLeaderboardPublic', value)}
          />
        </div>

        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={handleSave}
            disabled={saving}
            className="rounded bg-sky-600 px-5 py-2 font-semibold text-white transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Visibility'}
          </button>
        </div>
      </div>
    </div>
  );
}

export default ReportingVisibilityEditor;
