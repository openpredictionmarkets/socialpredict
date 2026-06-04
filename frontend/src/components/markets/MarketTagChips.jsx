import React from 'react';

const colorClasses = {
  sky: 'border-sky-500/40 bg-sky-950/40 text-sky-100',
  cyan: 'border-cyan-500/40 bg-cyan-950/40 text-cyan-100',
  teal: 'border-teal-500/40 bg-teal-950/40 text-teal-100',
  emerald: 'border-emerald-500/40 bg-emerald-950/40 text-emerald-100',
  lime: 'border-lime-500/40 bg-lime-950/40 text-lime-100',
  amber: 'border-amber-500/40 bg-amber-950/40 text-amber-100',
  orange: 'border-orange-500/40 bg-orange-950/40 text-orange-100',
  rose: 'border-rose-500/40 bg-rose-950/40 text-rose-100',
  violet: 'border-violet-500/40 bg-violet-950/40 text-violet-100',
  slate: 'border-gray-500/40 bg-gray-800/70 text-gray-200',
};

export const MARKET_TAG_COLOR_OPTIONS = [
  { key: 'slate', label: 'Slate', guidance: 'Default / general' },
  { key: 'sky', label: 'Sky', guidance: 'Open questions / civic' },
  { key: 'cyan', label: 'Cyan', guidance: 'Science / technology' },
  { key: 'teal', label: 'Teal', guidance: 'Health / environment' },
  { key: 'emerald', label: 'Emerald', guidance: 'Finance / growth' },
  { key: 'lime', label: 'Lime', guidance: 'Sports / active events' },
  { key: 'amber', label: 'Amber', guidance: 'Warnings / high attention' },
  { key: 'orange', label: 'Orange', guidance: 'Culture / media' },
  { key: 'rose', label: 'Rose', guidance: 'Politics / conflict' },
  { key: 'violet', label: 'Violet', guidance: 'Meta / platform' },
];

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
