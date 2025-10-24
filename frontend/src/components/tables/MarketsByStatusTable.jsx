import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from './MobileMarketCard';
import LoadingSpinner from '../loaders/LoadingSpinner';
import ExpandableLink from '../utils/ExpandableLink';
import { getResolvedText, getResultCssClass } from '../../utils/labelMapping';

const DEFAULT_LIMIT = 50;
const DEFAULT_CREATOR_EMOJI = '👤';

const toNumber = (value, fallback = 0) => {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : fallback;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const normalizeMarketOverview = (raw) => {
  if (!raw || typeof raw !== 'object') {
    return null;
  }

  if (raw.market || raw.creator) {
    return {
      market: raw.market ?? {},
      creator: raw.creator ?? {
        username: raw.market?.creatorUsername ?? 'unknown',
        personalEmoji: DEFAULT_CREATOR_EMOJI,
      },
      lastProbability: toNumber(raw.lastProbability),
      numUsers: toNumber(raw.numUsers),
      totalVolume: toNumber(raw.totalVolume),
      marketDust: toNumber(raw.marketDust),
    };
  }

  const market = raw;

  return {
    market,
    creator: {
      username: market.creatorUsername ?? 'unknown',
      personalEmoji: market.personalEmoji ?? DEFAULT_CREATOR_EMOJI,
      displayName: market.displayName,
    },
    lastProbability: toNumber(market.lastProbability),
    numUsers: toNumber(market.numUsers),
    totalVolume: toNumber(market.totalVolume),
    marketDust: toNumber(market.marketDust),
  };
};

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

const MarketRow = ({ marketData }) => {
  const market = marketData?.market ?? {};
  const marketId = market.id ?? market.marketId;
  const creator = marketData?.creator ?? {};
  const creatorUsername = creator.username ?? market.creatorUsername ?? 'unknown';
  const creatorEmoji = creator.personalEmoji ?? DEFAULT_CREATOR_EMOJI;
  const probability = toNumber(marketData?.lastProbability);
  const probabilityDisplay = Number.isFinite(probability)
    ? probability.toFixed(3)
    : '—';
  const numUsers = toNumber(marketData?.numUsers);
  const totalVolume = toNumber(marketData?.totalVolume);
  const resolutionDate = market?.resolutionDateTime;
  const questionTitle = market?.questionTitle ?? 'Untitled market';
  const isResolved = typeof market?.isResolved === 'boolean'
    ? market.isResolved
    : (typeof market?.status === 'string' && market.status.toLowerCase() === 'resolved');
  const resolutionResult = market?.resolutionResult ?? market?.status ?? '';

  return (
    <tr className='hover:bg-gray-700 transition-colors duration-200'>
      <td className='px-6 py-4 whitespace-nowrap'>
        <Link
          to={marketId ? `/markets/${marketId}` : '#'}
          className='text-blue-400 hover:text-blue-300'
        >
          ⬆️⬇️
        </Link>
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-300'>
        {probabilityDisplay}
      </td>
      <td className='px-6 py-4 text-sm font-medium text-gray-300'>
        <ExpandableLink
          text={questionTitle}
          to={marketId ? `/markets/${marketId}` : '#'}
          maxLength={45}
          className=''
          linkClassName='hover:text-blue-400 transition-colors duration-200'
          buttonClassName='text-xs text-blue-400 hover:text-blue-300 transition-colors ml-1'
          expandIcon='📐'
        />
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        {formatResolutionDate(resolutionDate)}
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        <Link
          to={`/user/${creatorUsername}`}
          className='flex items-center hover:text-blue-400 transition-colors duration-200'
        >
          <span role='img' aria-label='Creator' className='mr-2'>
            {creatorEmoji}
          </span>
          @{creatorUsername}
        </Link>
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        {numUsers}
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        {totalVolume}
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>0</td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        {isResolved ? (
          <span className={getResultCssClass(resolutionResult)}>
            {getResolvedText(resolutionResult, market)}
          </span>
        ) : (
          'Pending'
        )}
      </td>
    </tr>
  );
};

function MarketsByStatusTable({ status }) {
  const [marketsData, setMarketsData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const controller = new AbortController();

    const fetchMarkets = async () => {
      setLoading(true);
      setError('');

      try {
        const url = new URL(`${API_URL}/v0/markets`);
        const params = new URLSearchParams();

        if (status && status.toLowerCase() !== 'all') {
          params.set('status', status.toUpperCase());
        }

        params.set('limit', String(DEFAULT_LIMIT));
        url.search = params.toString();

        const response = await fetch(url.toString(), { signal: controller.signal });

        if (!response.ok) {
          throw new Error(`Failed to fetch ${status} markets`);
        }

        const data = await response.json();
        const rawMarkets = Array.isArray(data.markets) ? data.markets : [];
        const normalized = rawMarkets
          .map(normalizeMarketOverview)
          .filter((item) => item !== null);

        setMarketsData(normalized);
      } catch (err) {
        if (err.name === 'AbortError') {
          return;
        }
        console.error(`Error fetching ${status} market data:`, err);
        setError(err.message || String(err));
      } finally {
        setTimeout(() => setLoading(false), 300);
      }
    };

    fetchMarkets();

    return () => controller.abort();
  }, [status]);

  if (loading)
    return (
      <div className='p-4 text-center'>
        <LoadingSpinner />
        Loading {status} markets...
      </div>
    );

  if (error)
    return <div className='p-4 text-center text-red-500'>Error: {error}</div>;

  return (
    <div className='w-full'>
      {marketsData.length === 0 ? (
        <div className='p-4 text-center text-gray-400'>
          No {status} markets found.
        </div>
      ) : (
        <>
          <div className='md:hidden'>
            {marketsData.map((marketData, index) => (
              <MobileMarketCard key={index} marketData={marketData} />
            ))}
          </div>
          <div className='hidden md:block bg-gray-800 shadow-md rounded-lg overflow-hidden'>
            <div className='overflow-x-auto'>
              <table className='min-w-full divide-y divide-gray-700'>
                <TableHeader />
                <tbody className='bg-gray-800 divide-y divide-gray-700'>
                  {marketsData.map((marketData, index) => (
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

export default MarketsByStatusTable;
