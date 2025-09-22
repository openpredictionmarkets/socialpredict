import React from 'react';

export default function TradeCTA({ onClick, disabled }) {
  return (
    <div
      className="md:hidden fixed inset-x-0 bottom-0 z-50 bg-slate-900/80 backdrop-blur p-3"
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      data-testid="mobile-trade-cta"
    >
      <button
        type="button"
        onClick={onClick}
        disabled={disabled}
        className="w-full rounded-xl px-6 py-3 text-base font-semibold shadow-lg focus:outline-none focus:ring focus:ring-slate-500 disabled:opacity-50 bg-blue-600 hover:bg-blue-700 text-white"
      >
        TRADE
      </button>
    </div>
  );
}
