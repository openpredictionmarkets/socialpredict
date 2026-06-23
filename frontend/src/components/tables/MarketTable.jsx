import React from 'react';
import { Link } from 'react-router-dom';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from './MobileMarketCard';
import ExpandableText from '../utils/ExpandableText';
import { getResolvedText, getResultCssClass } from '../../utils/labelMapping';
import StewardTag, { stewardUsernameFor } from '../markets/StewardTag';
import MarketTagChips from '../markets/MarketTagChips';
import {
  groupMarketRows,
  groupedMarketBadgeLabel,
  groupedMarketResolutionSummary,
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

const MarketResolutionCell = ({ marketData }) => {
  const groupedSummary = isGroupedMarketAggregate(marketData)
    ? groupedMarketResolutionSummary(marketData)
    : null;
  if (groupedSummary) {
    return <span className={groupedSummary.className}>{groupedSummary.label}</span>;
  }

  return marketData.market.isResolved ? (
    <span className={getResultCssClass(marketData.market.resolutionResult)}>
      {getResolvedText(marketData.market.resolutionResult, marketData.market)}
    </span>
  ) : (
    'Pending'
  );
};

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
    <td className='px-6 py-4 text-sm font-medium text-gray-300'>
      <Link
        to={marketDisplayRoute(marketData)}
        className='hover:text-blue-400 transition-colors duration-200 block max-w-xs'
        title={marketData.market.questionTitle}
      >
        <ExpandableText
          text={marketData.market.questionTitle}
          maxLength={45}
          className=""
          expandedClassName="mt-2 p-2 bg-gray-700 rounded border border-gray-600 relative z-10"
          buttonClassName="text-xs text-blue-400 hover:text-blue-300 transition-colors ml-1"
          showFullTextInExpanded={true}
          expandIcon="📐"
        />
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
      <MarketResolutionCell marketData={marketData} />
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
  const groupedMarkets = groupMarketRows(markets);

  return (
    <>
      {/* Mobile view */}
      <div className='md:hidden'>
        {groupedMarkets.map((marketData, index) => (
          <MobileMarketCard key={index} marketData={marketData} />
        ))}
      </div>
      
      {/* Desktop view */}
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
  );
};

export default MarketTable;
