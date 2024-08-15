import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatDateTimeForGrid } from '../utils/dateTimeTools/FormDateTimeTools';
import MarketChart from '../charts/MarketChart';
import ActivityTabs from '../../components/tabs/ActivityTabs';

function MarketDetailsTable({
  market,
  creator,
  numUsers,
  totalVolume,
  currentProbability,
  probabilityChanges,
  marketId,
}) {
  const [showFullDescription, setShowFullDescription] = useState(false);

  return (
    <div className='bg-gray-900 text-gray-300 p-4 rounded-lg shadow-lg w-full overflow-hidden'>
      <div className='mb-4'>
        <h1 className='text-xl font-semibold text-white mb-2 break-words'>
          {market.questionTitle}
        </h1>
        <div className='flex items-center space-x-2 text-sm text-gray-400'>
          <Link
            to={`/user/${market.creatorUsername}`}
            className='hover:text-blue-400 transition-colors duration-200'
          >
            <span role='img' aria-label='Creator'>
              {creator.personalEmoji}
            </span>
            @{market.creatorUsername}
          </Link>
          <span>•</span>
          <span>🪙 {currentProbability.toFixed(2)}</span>
        </div>
      </div>

      <div className='mb-4'>
        <MarketChart
          data={probabilityChanges}
          currentProbability={currentProbability}
          title='Probability Changes'
          className='w-full'
        />
      </div>

      <div className='mb-4'>
        <button
          onClick={() => setShowFullDescription(!showFullDescription)}
          className='w-full py-2 bg-gray-700 hover:bg-gray-600 transition-colors duration-200 rounded-lg text-center text-sm'
        >
          {showFullDescription ? 'Hide Description' : 'Show Full Description'}
        </button>
      </div>

      <div className='mb-4 bg-gray-800 p-4 rounded-lg'>
        <p
          className={`text-sm break-words whitespace-pre-wrap ${
            showFullDescription
              ? ''
              : 'sm:max-h-24 h-16 overflow-y-auto sm:overflow-hidden'
          }`}
          style={{
            wordBreak: 'break-word',
            overflowWrap: 'break-word',
            hyphens: 'auto',
          }}
        >
          {market.description}
        </p>
      </div>

      <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
        <div className='bg-gray-800 p-2 rounded-lg'>
          <div className='text-xs text-gray-400'>Users</div>
          <div className='text-sm font-semibold'>👤 {numUsers}</div>
        </div>
        <div className='bg-gray-800 p-2 rounded-lg'>
          <div className='text-xs text-gray-400'>Volume</div>
          <div className='text-sm font-semibold'>
            📊 {totalVolume.toFixed(2)}
          </div>
        </div>
        <div className='bg-gray-800 p-2 rounded-lg'>
          <div className='text-xs text-gray-400'>Comments</div>
          <div className='text-sm font-semibold'>💬 0</div>
        </div>
        <div className='bg-gray-800 p-2 rounded-lg'>
          <div className='text-xs text-gray-400'>Closes</div>
          <div className='text-sm font-semibold'>
            📅{' '}
            {formatDateTimeForGrid(market.resolutionDateTime).toLocaleString()}
          </div>
        </div>
      </div>

      <div className='mx-auto mt-8 w-full'>
        <ActivityTabs marketId={marketId} />
      </div>
    </div>
  );
}

export default MarketDetailsTable;
