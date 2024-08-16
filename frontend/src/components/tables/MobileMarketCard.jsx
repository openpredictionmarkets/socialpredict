import React from 'react';
import { Link } from 'react-router-dom';
import formatResolutionDate from '../../helpers/formatResolutionDate';

const MobileMarketCard = ({ marketData }) => {
  const isMarketOpen =
    !marketData.market.isResolved &&
    formatResolutionDate(marketData.market.resolutionDateTime) !== 'Closed';

  return (
    <div className='bg-gray-800 p-4 mb-4 rounded-lg'>
      <div className='grid grid-cols-[1fr,auto] gap-2 items-center mb-2'>
        <Link
          to={`/user/${marketData.creator.username}`}
          className='text-gray-400 hover:text-blue-400 transition-colors duration-200 truncate'
        >
          <span>
            {marketData.creator.username} {marketData.creator.personalEmoji}
          </span>
        </Link>
        {isMarketOpen ? (
          <Link
            to={`/markets/${marketData.market.id}`}
            className='bg-blue-500 text-white px-3 py-1 rounded-full text-sm whitespace-nowrap'
          >
            Bet
          </Link>
        ) : (
          <span className='bg-gray-500 text-white px-3 py-1 rounded-full text-sm whitespace-nowrap'>
            Closed
          </span>
        )}
      </div>
      <Link
        to={`/markets/${marketData.market.id}`}
        className='text-blue-400 hover:text-blue-300 font-medium block mb-2 truncate'
      >
        {marketData.market.questionTitle}
      </Link>
      <div className='grid grid-cols-3 text-sm text-gray-400'>
        <span className='truncate'>👤 {marketData.numUsers}</span>
        <span className='text-center'>
          {marketData.lastProbability.toFixed(2)}%
        </span>
        <span
          className={`text-right ${
            marketData.market.isResolved
              ? marketData.market.resolutionResult === 'YES'
                ? 'text-green-400'
                : 'text-red-400'
              : 'text-green-400'
          }`}
        >
          {marketData.market.isResolved
            ? marketData.market.resolutionResult
            : 'Pending'}
        </span>
      </div>
    </div>
  );
};

export default MobileMarketCard;
