import React from 'react';
import { Link } from 'react-router-dom';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from './MobileMarketCard';

const TableHeader = () => (
  <thead className='bg-gray-900'>
    <tr>
      {[
        'Trade',
        '🪙',
        'Question',
        '📅 Closes',
        'Creator',
        '👤 Users',
        '📊 Size',
        '💬',
        'Resolution',
      ].map((header, index) => (
        <th
          key={index}
          className='px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider'
        >
          {header}
        </th>
      ))}
    </tr>
  </thead>
);

const MarketRow = ({ marketData }) => (
  <tr className='hover:bg-gray-700 transition-colors duration-200'>
    <td className='px-6 py-4 whitespace-nowrap'>
      <Link
        to={`/markets/${marketData.market.id}`}
        className='text-blue-400 hover:text-blue-300'
      >
        ⬆️⬇️
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-300'>
      {marketData.lastProbability.toFixed(3)}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-300'>
      <Link
        to={`/markets/${marketData.market.id}`}
        className='hover:text-blue-400 transition-colors duration-200 block max-w-xs overflow-hidden overflow-ellipsis'
        title={marketData.market.questionTitle}
      >
        {marketData.market.questionTitle}
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {formatResolutionDate(marketData.market.resolutionDateTime)}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      <Link
        to={`/user/${marketData.creator.username}`}
        className='flex items-center hover:text-blue-400 transition-colors duration-200'
      >
        <span role='img' aria-label='Creator' className='mr-2'>
          {marketData.creator.personalEmoji}
        </span>
        @{marketData.creator.username}
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.numUsers}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.totalVolume}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>0</td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.market.isResolved ? (
        <span
          className={
            marketData.market.resolutionResult === 'YES'
              ? 'text-green-400'
              : 'text-red-400'
          }
        >
          {marketData.market.resolutionResult === 'YES'
            ? 'Resolved YES'
            : 'Resolved NO'}
        </span>
      ) : (
        'Pending'
      )}
    </td>
  </tr>
);

// Pure table component that renders markets from props
const MarketTable = ({ markets }) => {
  // Handle empty or invalid markets data
  if (!markets || markets.length === 0) {
    return (
      <div className='p-4 text-center text-gray-400'>No markets found.</div>
    );
  }

  return (
    <>
      {/* Mobile view */}
      <div className='md:hidden'>
        {markets.map((marketData, index) => (
          <MobileMarketCard key={index} marketData={marketData} />
        ))}
      </div>
      
      {/* Desktop view */}
      <div className='hidden md:block bg-gray-800 shadow-md rounded-lg overflow-hidden'>
        <div className='overflow-x-auto'>
          <table className='min-w-full divide-y divide-gray-700'>
            <TableHeader />
            <tbody className='bg-gray-800 divide-y divide-gray-700'>
              {markets.map((marketData, index) => (
                <MarketRow key={index} marketData={marketData} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
};

export default MarketTable;
