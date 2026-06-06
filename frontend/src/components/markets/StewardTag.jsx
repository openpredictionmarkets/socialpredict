import React from 'react';
import { Link } from 'react-router-dom';

export const stewardUsernameFor = (market = {}, fallbackUsername = '') => (
  market.stewardUsername
  ?? market.StewardUsername
  ?? fallbackUsername
  ?? ''
);

const StewardTag = ({ username, creatorUsername = '', className = '' }) => {
  const normalizedUsername = String(username || '').trim();
  const normalizedCreatorUsername = String(creatorUsername || '').trim();
  if (!normalizedUsername || normalizedUsername === 'unknown') {
    return null;
  }
  if (normalizedCreatorUsername && normalizedUsername === normalizedCreatorUsername) {
    return null;
  }

  return (
    <Link
      to={`/user/${normalizedUsername}`}
      className={`inline-flex items-center gap-1 rounded-full border border-info-blue/30 bg-info-blue/10 px-2 py-0.5 text-xs font-medium text-custom-gray-verylight transition hover:border-info-blue/60 hover:bg-info-blue/15 ${className}`}
      title={`Market steward: ${normalizedUsername}`}
    >
      <span className="font-mono text-[0.62rem] uppercase tracking-[0.14em] text-info-blue">Steward</span>
      <span className="text-gray-300">@{normalizedUsername}</span>
    </Link>
  );
};

export default StewardTag;
