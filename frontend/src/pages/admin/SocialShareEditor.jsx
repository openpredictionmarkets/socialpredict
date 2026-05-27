import React, { useEffect, useMemo, useState } from 'react';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';

const descriptionGuidance = 'Aim for 110-160 characters. The backend allows up to 220 characters.';
const imageGuidance = 'Use a public 1200x630px image URL. Root-relative paths like /og/card.png are allowed.';

function SocialShareEditor() {
  const [settings, setSettings] = useState({
    siteName: '',
    defaultDescription: '',
    defaultImageUrl: '',
    imageAlt: '',
    version: 0,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchSettings();
  }, []);

  const descriptionLength = settings.defaultDescription.length;
  const descriptionTone = useMemo(() => {
    if (descriptionLength === 0) return 'text-gray-400';
    if (descriptionLength < 110 || descriptionLength > 160) return 'text-amber-300';
    return 'text-green-300';
  }, [descriptionLength]);

  const fetchSettings = async () => {
    try {
      const data = await apiRequest('/v0/content/social-share', {
        fallbackMessage: 'Failed to load social share settings',
      });
      setSettings({
        siteName: data.siteName || '',
        defaultDescription: data.defaultDescription || '',
        defaultImageUrl: data.defaultImageUrl || '',
        imageAlt: data.imageAlt || '',
        version: data.version || 0,
      });
    } catch (err) {
      setError(err.message || 'Error loading social share settings');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field, value) => {
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
      const data = await authenticatedApiRequest('/v0/admin/content/social-share', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings),
        fallbackMessage: 'Failed to save social share settings',
      });
      setSettings({
        siteName: data.siteName || '',
        defaultDescription: data.defaultDescription || '',
        defaultImageUrl: data.defaultImageUrl || '',
        imageAlt: data.imageAlt || '',
        version: data.version || 0,
      });
      setMessage('Social share settings saved successfully.');
    } catch (err) {
      setError(err.message || 'Error saving social share settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-primary-background p-8">
        <div className="max-w-4xl mx-auto">
          <p className="text-white">Loading social share settings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-primary-background p-8">
      <div className="max-w-5xl mx-auto space-y-6">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-blue-300">CMS</p>
          <h1 className="text-3xl font-bold text-white mt-2">Social Share Settings</h1>
          <p className="text-gray-300 mt-2 max-w-3xl">
            Configure the default Open Graph image, fallback description, and site name used by public market share cards.
          </p>
        </div>

        {message && (
          <div className="bg-green-600 text-white p-4 rounded">
            {message}
          </div>
        )}

        {error && (
          <div className="bg-red-600 text-white p-4 rounded">
            {error}
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
          <div className="bg-gray-800 rounded-lg p-6 space-y-6 border border-gray-700">
            <div>
              <label className="block text-white font-semibold mb-2">Site Name</label>
              <input
                type="text"
                maxLength={80}
                value={settings.siteName}
                onChange={(e) => handleInputChange('siteName', e.target.value)}
                className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="SocialPredict"
              />
            </div>

            <div>
              <div className="flex items-center justify-between gap-4 mb-2">
                <label className="block text-white font-semibold">Default Description</label>
                <span className={`text-sm ${descriptionTone}`}>{descriptionLength}/220</span>
              </div>
              <textarea
                maxLength={220}
                rows={4}
                value={settings.defaultDescription}
                onChange={(e) => handleInputChange('defaultDescription', e.target.value)}
                className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="Prediction markets for the social web"
              />
              <p className="text-gray-400 text-sm mt-2">{descriptionGuidance}</p>
            </div>

            <div>
              <label className="block text-white font-semibold mb-2">Default Open Graph Image URL</label>
              <input
                type="text"
                maxLength={500}
                value={settings.defaultImageUrl}
                onChange={(e) => handleInputChange('defaultImageUrl', e.target.value)}
                className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="/og/socialpredict-card.png"
              />
              <p className="text-gray-400 text-sm mt-2">{imageGuidance}</p>
            </div>

            <div>
              <label className="block text-white font-semibold mb-2">Image Alt Text</label>
              <input
                type="text"
                maxLength={160}
                value={settings.imageAlt}
                onChange={(e) => handleInputChange('imageAlt', e.target.value)}
                className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="SocialPredict share card"
              />
            </div>

            <div className="flex items-center justify-between pt-2">
              <span className="text-gray-400">Current Version: {settings.version}</span>
              <button
                onClick={handleSave}
                disabled={saving}
                className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-6 py-2 rounded font-semibold transition-colors"
              >
                {saving ? 'Saving...' : 'Save Settings'}
              </button>
            </div>
          </div>

          <aside className="bg-gray-900 border border-gray-700 rounded-lg p-5 h-fit">
            <h2 className="text-white font-semibold text-lg mb-4">Share Preview</h2>
            <div className="rounded-lg overflow-hidden border border-gray-700 bg-gray-950">
              <div className="aspect-[1200/630] bg-gray-800 flex items-center justify-center text-gray-400 text-sm">
                {settings.defaultImageUrl ? (
                  <img
                    src={settings.defaultImageUrl}
                    alt={settings.imageAlt || 'Social share preview'}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  'No default image URL'
                )}
              </div>
              <div className="p-4">
                <p className="text-xs uppercase tracking-wide text-gray-500">Public market card</p>
                <h3 className="text-white font-semibold mt-1">Example Market | {settings.siteName || 'SocialPredict'}</h3>
                <p className="text-gray-300 text-sm mt-2">
                  {settings.defaultDescription || 'Prediction markets for the social web'}
                </p>
              </div>
            </div>
            <p className="text-gray-400 text-sm mt-4">
              Market pages still use the market title and description when present. These settings provide the site name, fallback description, and default image.
            </p>
          </aside>
        </div>
      </div>
    </div>
  );
}

export default SocialShareEditor;
