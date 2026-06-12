import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from '../../components/tables/MobileMarketCard';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { getResolvedText, getResultCssClass } from '../../utils/labelMapping';
import StewardTag, { stewardUsernameFor } from '../markets/StewardTag';
import MarketTagChips from '../markets/MarketTagChips';
import {
  groupMarketRows,
  groupedMarketBadgeLabel,
  isGroupedMarketAggregate,
  marketDisplayRoute,
  marketProbabilityDisplay,
} from '../../helpers/marketGroups';

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
        to={marketDisplayRoute(marketData)}
        className='text-blue-400 hover:text-blue-300'
      >
        ⬆️⬇️
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-300'>
      {marketProbabilityDisplay(marketData)}
    </td>
    <td className='px-6 py-4 whitespace-normal text-sm font-medium text-gray-300'>
      <Link
        to={marketDisplayRoute(marketData)}
        className='hover:text-blue-400 transition-colors duration-200 block max-w-xs overflow-hidden overflow-ellipsis'
        title={marketData.market.questionTitle}
      >
        {marketData.market.questionTitle}
      </Link>
      <div className='mt-2 flex flex-wrap items-center gap-2'>
        {isGroupedMarketAggregate(marketData) && (
          <span className='rounded-full border border-cyan-500/40 bg-cyan-950/40 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-cyan-100'>
            {groupedMarketBadgeLabel(marketData)}
          </span>
        )}
        <MarketTagChips tags={marketData.market.tags || []} />
      </div>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {formatResolutionDate(marketData.market.resolutionDateTime)}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      <div className='flex flex-col items-start gap-2'>
        <Link
          to={`/user/${marketData.creator.username}`}
          className='flex items-center hover:text-blue-400 transition-colors duration-200'
        >
          <span role='img' aria-label='Creator' className='mr-2'>
            {marketData.creator.personalEmoji}
          </span>
          @{marketData.creator.username}
        </Link>
        <StewardTag
          username={stewardUsernameFor(marketData.market, marketData.creator.username)}
          creatorUsername={marketData.creator.username}
        />
      </div>
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
        <span className={getResultCssClass(marketData.market.resolutionResult)}>
          {getResolvedText(marketData.market.resolutionResult, marketData.market)}
        </span>
      ) : (
        'Pending'
      )}
    </td>
  </tr>
);

function MarketsTable() {
  const [marketsData, setMarketsData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchMarkets = async () => {
      try {
        const response = await fetch(`${API_URL}/v0/markets`);
        if (!response.ok) throw new Error('Failed to fetch markets');
        const data = await response.json();
        setMarketsData(data.markets || []);
      } catch (error) {
        console.error('Error fetching market data:', error);
        setError(error.toString());
      } finally {
        setTimeout(() => setLoading(false), 300);
      }
    };

    fetchMarkets();
  }, []);

  if (loading)
    return (
      <div className='p-4 text-center'>
        <LoadingSpinner />
        Loading markets...
      </div>
    );
  if (error)
    return <div className='p-4 text-center text-red-500'>Error: {error}</div>;

  const groupedMarkets = groupMarketRows(marketsData);

  return (
    <div className='w-full md:w-full h-[calc(100vh-40px)] sm:h-full overflow-y-auto px-4 md:px-6 lg:px-8'>
      <h1 className='text-2xl font-semibold text-gray-300 mb-6'>Markets</h1>
      {marketsData.length === 0 ? (
        <div className='p-4 text-center text-gray-400'>No markets found.</div>
      ) : (
        <>
          <div className='md:hidden'>
            {groupedMarkets.map((marketData, index) => (
              <MobileMarketCard key={index} marketData={marketData} />
            ))}
          </div>
          <div className='hidden md:block bg-gray-800 shadow-md rounded-lg overflow-hidden'>
            <div className='overflow-x-auto'>
              <table className='min-w-full divide-y divide-gray-700'>
                <TableHeader />
                <tbody className='bg-gray-800 divide-y divide-gray-700'>
                  {groupedMarkets.map((marketData, index) => (
                    <MarketRow key={index} marketData={marketData} />
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default MarketsTable;
