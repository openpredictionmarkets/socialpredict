import React from 'react';

const colorClasses = {
  sky: 'border-sky-500/40 bg-sky-950/40 text-sky-100',
  emerald: 'border-emerald-500/40 bg-emerald-950/40 text-emerald-100',
  amber: 'border-amber-500/40 bg-amber-950/40 text-amber-100',
  rose: 'border-rose-500/40 bg-rose-950/40 text-rose-100',
  slate: 'border-gray-500/40 bg-gray-800/70 text-gray-200',
};

const chipClassFor = (tag) => {
  const key = String(tag?.colorKey || 'slate').toLowerCase();
  return colorClasses[key] || colorClasses.slate;
};

const MarketTagChips = ({ tags = [], className = '' }) => {
  const visibleTags = (tags || []).filter((tag) => tag?.slug || tag?.displayName);
  if (!visibleTags.length) {
    return null;
  }

  return (
    <div className={`flex flex-wrap gap-2 ${className}`.trim()} aria-label="Market tags">
      {visibleTags.map((tag) => (
        <span
          key={tag.slug || tag.displayName}
          className={`rounded-full border px-2.5 py-1 text-xs font-semibold ${chipClassFor(tag)}`}
          title={tag.description || tag.displayName || tag.slug}
        >
          {tag.displayName || tag.slug}
        </span>
      ))}
    </div>
  );
};

export default MarketTagChips;
