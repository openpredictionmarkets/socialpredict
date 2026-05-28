import React, { useEffect, useMemo, useState } from 'react';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';
import { API_URL } from '../../config';

const descriptionGuidance = 'Aim for 110-160 characters. The backend allows up to 220 characters.';
const imageGuidance = 'Upload a PNG, JPEG, or WebP up to 5 MB. Recommended dimensions are 1200x630px.';
const hostedImageGuidance = 'For high-traffic deployments, use a hosted image URL/CDN so social crawlers do not add avoidable load to this server.';
const adminMutationReasonMessages = {
  PASSWORD_CHANGE_REQUIRED: 'Your admin password must be changed before updating CMS social share settings.',
  AUTHORIZATION_DENIED: 'Only admin users can update CMS social share settings.',
  INVALID_TOKEN: 'Your session expired. Please log in again.',
};

const resolvePreviewImageUrl = (imageUrl) => {
  if (!imageUrl) return '';
  if (/^https?:\/\//.test(imageUrl)) return imageUrl;

  const normalizedPath = imageUrl.startsWith('/') ? imageUrl : `/${imageUrl}`;
  if (normalizedPath.startsWith('/v0/')) {
    return `${API_URL}${normalizedPath}`;
  }

  const frontendOrigin = typeof window !== 'undefined' ? window.location.origin : '';
  return `${frontendOrigin}${normalizedPath}`;
};

const appendPreviewCacheBust = (imageUrl, cacheKey) => {
  if (!imageUrl || !cacheKey) return imageUrl;
  const separator = imageUrl.includes('?') ? '&' : '?';
  return `${imageUrl}${separator}previewVersion=${encodeURIComponent(cacheKey)}`;
};

const normalizeSettings = (data = {}) => ({
  siteName: data.siteName || '',
  defaultDescription: data.defaultDescription || '',
  defaultImageUrl: data.defaultImageUrl || '',
  imageEnabled: data.imageEnabled !== false,
  imageAlt: data.imageAlt || '',
  updatedAt: data.updatedAt || '',
  version: data.version || 0,
});

function SocialShareEditor() {
  const [settings, setSettings] = useState(normalizeSettings());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
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
  const previewCacheKey = settings.updatedAt || settings.version;
  const previewImageUrl = useMemo(() => {
    const resolvedUrl = resolvePreviewImageUrl(settings.defaultImageUrl);
    return appendPreviewCacheBust(resolvedUrl, previewCacheKey);
  }, [settings.defaultImageUrl, previewCacheKey]);

  const fetchSettings = async () => {
    try {
      const data = await apiRequest('/v0/content/social-share', {
        fallbackMessage: 'Failed to load social share settings',
      });
      setSettings(normalizeSettings(data));
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
        reasonMessages: adminMutationReasonMessages,
        fallbackMessage: 'Failed to save social share settings',
      });
      setSettings(normalizeSettings(data));
      setMessage('Social share settings saved successfully.');
    } catch (err) {
      setError(err.message || 'Error saving social share settings');
    } finally {
      setSaving(false);
    }
  };

  const handleImageUpload = async (event) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setUploading(true);
    setMessage('');
    setError('');

    try {
      const formData = new FormData();
      formData.append('image', file);
      formData.append('imageAlt', settings.imageAlt || '');

      const data = await authenticatedApiRequest('/v0/admin/content/social-share/image', {
        method: 'POST',
        body: formData,
        reasonMessages: adminMutationReasonMessages,
        fallbackMessage: 'Failed to upload social share image',
      });

      setSettings(normalizeSettings(data));
      setMessage('Social share image uploaded successfully.');
    } catch (err) {
      setError(err.message || 'Error uploading social share image');
    } finally {
      setUploading(false);
      event.target.value = '';
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
              <label className="block text-white font-semibold mb-2">Default Open Graph Image</label>
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp"
                onChange={handleImageUpload}
                disabled={uploading}
                className="block w-full text-sm text-gray-300 file:mr-4 file:rounded file:border-0 file:bg-blue-600 file:px-4 file:py-2 file:font-semibold file:text-white hover:file:bg-blue-700 disabled:opacity-50"
              />
              <p className="text-gray-400 text-sm mt-2">
                {uploading ? 'Uploading image...' : imageGuidance}
              </p>
              <label className="flex items-start gap-3 mt-4 rounded border border-amber-500/40 bg-amber-950/40 p-3 text-sm text-amber-100">
                <input
                  type="checkbox"
                  checked={settings.imageEnabled}
                  onChange={(e) => handleInputChange('imageEnabled', e.target.checked)}
                  className="mt-1 h-4 w-4"
                />
                <span>
                  <span className="block font-semibold text-amber-50">Include image metadata on shared market pages</span>
                  <span className="block mt-1">
                    Turn this off if public link previews are creating too much image traffic. Market pages will still share title, description, and URL metadata.
                  </span>
                </span>
              </label>
              <label className="block text-white font-semibold mt-4 mb-2">Current Image URL</label>
              <input
                type="text"
                maxLength={500}
                value={settings.defaultImageUrl}
                onChange={(e) => handleInputChange('defaultImageUrl', e.target.value)}
                className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="/og/socialpredict-share.png"
              />
              <p className="text-amber-200 text-sm mt-2">
                {hostedImageGuidance}
              </p>
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
                {settings.imageEnabled && previewImageUrl ? (
                  <img
                    key={previewImageUrl}
                    src={previewImageUrl}
                    alt={settings.imageAlt || 'Social share preview'}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  settings.imageEnabled ? 'No default image URL' : 'Image metadata disabled'
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
