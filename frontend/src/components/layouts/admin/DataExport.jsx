import React, { useState } from 'react';
import { API_URL } from '../../../config';

const EXPORTS = [
  {
    key: 'bets',
    label: 'Bets',
    description: 'All bet transactions (buy/sell, user, market, amount, timestamp)',
    url: `${API_URL}/v0/admin/export/bets`,
  },
  {
    key: 'markets',
    label: 'Markets',
    description: 'All markets (title, creator, resolution status, dates)',
    url: `${API_URL}/v0/admin/export/markets`,
  },
  {
    key: 'users',
    label: 'Users',
    description: 'All users — public data only, no passwords',
    url: `${API_URL}/v0/admin/export/users`,
  },
];

const DataExport = () => {
  const [loading, setLoading] = useState({});
  const [errors, setErrors] = useState({});

  const token = localStorage.getItem('token');

  const handleExport = async (exportDef) => {
    setLoading((prev) => ({ ...prev, [exportDef.key]: true }));
    setErrors((prev) => ({ ...prev, [exportDef.key]: null }));

    try {
      const response = await fetch(exportDef.url, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!response.ok) {
        const msg = await response.text();
        throw new Error(msg || `HTTP ${response.status}`);
      }

      // Derive filename from Content-Disposition or fallback
      const disposition = response.headers.get('Content-Disposition') || '';
      const match = disposition.match(/filename="([^"]+)"/);
      const filename = match ? match[1] : `${exportDef.key}_export.csv`;

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setErrors((prev) => ({ ...prev, [exportDef.key]: err.message }));
    } finally {
      setLoading((prev) => ({ ...prev, [exportDef.key]: false }));
    }
  };

  return (
    <div className='p-6 space-y-4'>
      <h2 className='text-xl font-semibold text-gray-200 mb-2'>Data Export</h2>
      <p className='text-gray-400 text-sm mb-4'>
        Download a point-in-time CSV snapshot of platform data.
      </p>

      {EXPORTS.map((exp) => (
        <div
          key={exp.key}
          className='flex items-center justify-between bg-gray-700 rounded-lg p-4'
        >
          <div>
            <p className='text-gray-200 font-medium'>{exp.label}</p>
            <p className='text-gray-400 text-sm mt-0.5'>{exp.description}</p>
            {errors[exp.key] && (
              <p className='text-red-400 text-xs mt-1'>{errors[exp.key]}</p>
            )}
          </div>
          <button
            onClick={() => handleExport(exp)}
            disabled={loading[exp.key]}
            className='ml-4 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg
                       hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors shrink-0'
          >
            {loading[exp.key] ? 'Downloading…' : 'Download CSV'}
          </button>
        </div>
      ))}
    </div>
  );
};

export default DataExport;
