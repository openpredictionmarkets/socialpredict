import React from 'react';
import { getResolvedText, getResultCssClass } from '../../utils/labelMapping';

const ResolutionAlert = ({ isResolved, resolutionResult, market }) => {
  if (!isResolved) return null;

  const isYes = resolutionResult.toUpperCase() === 'YES';
  const bgColor = isYes ? 'bg-blue-900/30' : 'bg-red-900/30';
  const borderColor = isYes ? 'border-blue-500/50' : 'border-red-500/50';
  const textColor = isYes ? 'text-blue-200' : 'text-red-200';
  const iconColor = isYes ? 'text-blue-400' : 'text-red-400';
  const strongColor = isYes ? 'text-green-300' : 'text-red-300';

  return (
    <div
      className={`mb-4 p-4 ${bgColor} border ${borderColor} rounded-lg ${textColor}`}
    >
      <div className='flex items-center'>
        <svg
          className={`h-5 w-5 ${iconColor} mr-2`}
          fill='none'
          viewBox='0 0 24 24'
          stroke='currentColor'
        >
          <path
            strokeLinecap='round'
            strokeLinejoin='round'
            strokeWidth={2}
            d='M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
          />
        </svg>
        <div>
          <p className='text-sm font-medium'>Market Resolved</p>
          <p className='text-sm'>
            This market has been resolved as{' '}
            <strong className={strongColor}>
              {market ? getResolvedText(resolutionResult, market).replace('Resolved ', '') : resolutionResult}
            </strong>
          </p>
        </div>
      </div>
    </div>
  );
};

export default ResolutionAlert;
