import React, { useEffect, useState } from 'react';
import {API_URL} from "../../config";

function HomeEditor() {
  const [content, setContent] = useState({
    title: '',
    html: '',
    version: 0
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchContent();
  }, []);

  const fetchContent = async () => {
    try {
      const response = await fetch(`${API_URL}/v0/content/home`);
      if (response.ok) {
        const data = await response.json();
        setContent({
          title: data.title || '',
          html: data.html || '',
          version: data.version || 0
        });
      } else {
        setError('Failed to load content');
      }
    } catch (err) {
      setError('Error loading content: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage('');
    setError('');

    try {
      const token = localStorage.getItem('token'); // Use correct token key from localStorage
      const response = await fetch(`${API_URL}/v0/admin/content/home`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          title: content.title,
          format: 'html',
          html: content.html,
          version: content.version
        })
      });

      if (response.ok) {
        const data = await response.json();
        setContent(prev => ({
          ...prev,
          version: data.version,
          html: data.html
        }));
        setMessage('Content saved successfully!');
      } else {
        const errorText = await response.text();
        setError('Failed to save: ' + errorText);
      }
    } catch (err) {
      setError('Error saving content: ' + err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field, value) => {
    setContent(prev => ({
      ...prev,
      [field]: value
    }));
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-primary-background p-8">
        <div className="max-w-4xl mx-auto">
          <p className="text-white">Loading homepage content...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-primary-background p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold text-white mb-8">Edit Homepage</h1>

        {message && (
          <div className="bg-green-600 text-white p-4 rounded mb-4">
            {message}
          </div>
        )}

        {error && (
          <div className="bg-red-600 text-white p-4 rounded mb-4">
            {error}
          </div>
        )}

        <div className="bg-gray-800 rounded-lg p-6 space-y-6">
          {/* Title Field */}
          <div>
            <label className="block text-white font-semibold mb-2">
              Title
            </label>
            <input
              type="text"
              value={content.title}
              onChange={(e) => handleInputChange('title', e.target.value)}
              className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
              placeholder="Homepage title"
            />
          </div>

          {/* Content Editor */}
          <div>
            <label className="block text-white font-semibold mb-2">
              HTML Content
            </label>
            <textarea
              value={content.html}
              onChange={(e) => handleInputChange('html', e.target.value)}
              rows={20}
              className="w-full p-3 bg-gray-700 text-white rounded border border-gray-600 focus:border-blue-500 focus:outline-none font-mono text-sm"
              placeholder="Enter HTML content..."
            />
          </div>

          {/* Version Info */}
          <div className="flex items-center justify-between">
            <span className="text-gray-400">
              Current Version: {content.version}
            </span>
            <button
              onClick={handleSave}
              disabled={saving}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-6 py-2 rounded font-semibold transition-colors"
            >
              {saving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>

          {/* Preview Section */}
          {content.html && (
            <div>
              <h3 className="text-white font-semibold mb-2">Preview:</h3>
              <div
                className="bg-gray-900 p-4 rounded border prose prose-invert max-w-none"
                dangerouslySetInnerHTML={{ __html: content.html }}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default HomeEditor;
