import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from './MobileMarketCard';
import LoadingSpinner from '../loaders/LoadingSpinner';
import ExpandableLink from '../utils/ExpandableLink';
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

const DEFAULT_LIMIT = 20;
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
  const route = marketDisplayRoute(marketData);
  const creator = marketData?.creator ?? {};
  const creatorUsername = creator.username ?? market.creatorUsername ?? 'unknown';
  const creatorEmoji = creator.personalEmoji ?? DEFAULT_CREATOR_EMOJI;
  const stewardUsername = stewardUsernameFor(market, creatorUsername);
  const probabilityDisplay = marketProbabilityDisplay(marketData);
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
          to={route}
          className='text-blue-400 hover:text-blue-300'
        >
          ⬆️⬇️
        </Link>
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-300'>
        {probabilityDisplay}
      </td>
      <td className='px-6 py-4 text-sm font-medium text-gray-300'>
        <div className='flex max-w-md flex-wrap items-center gap-2'>
          <ExpandableLink
            text={questionTitle}
            to={route}
            maxLength={45}
            className=''
            linkClassName='hover:text-blue-400 transition-colors duration-200'
            buttonClassName='text-xs text-blue-400 hover:text-blue-300 transition-colors ml-1'
            expandIcon='📐'
          />
          {isGroupedMarketAggregate(marketData) && (
            <span className='rounded-full border border-cyan-500/40 bg-cyan-950/40 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-cyan-100'>
              {groupedMarketBadgeLabel(marketData)}
            </span>
          )}
          <MarketTagChips tags={market.tags || []} />
        </div>
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        {formatResolutionDate(resolutionDate)}
      </td>
      <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
        <div className='flex flex-col items-start gap-2'>
          <Link
            to={`/user/${creatorUsername}`}
            className='flex items-center hover:text-blue-400 transition-colors duration-200'
          >
            <span role='img' aria-label='Creator' className='mr-2'>
              {creatorEmoji}
            </span>
            @{creatorUsername}
          </Link>
          <StewardTag username={stewardUsername} creatorUsername={creatorUsername} />
        </div>
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

function MarketsByStatusTable({ status, limit = DEFAULT_LIMIT, tagSlug = '', discoverySlug = 'markets' }) {
  const [marketsData, setMarketsData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [hasMore, setHasMore] = useState(true);
  const sentinelRef = useRef(null);
  const observerRef = useRef(null);
  const inFlightRef = useRef(false);
  const activeRequestRef = useRef(null);
  const nextOffsetRef = useRef(0);
  const pageSize = limit || DEFAULT_LIMIT;

  const fetchMarketsPage = useCallback(async ({ offset, append, signal }) => {
    if (append && inFlightRef.current) {
      return;
    }
    const requestId = Symbol('markets-page-request');
    activeRequestRef.current = requestId;
    inFlightRef.current = true;

    if (append) {
      setLoadingMore(true);
    } else {
      setLoading(true);
    }
    setError('');

    try {
      const safeDiscoverySlug = encodeURIComponent(discoverySlug || 'markets');
      const url = new URL(`${API_URL}/v0/read/market-discovery/${safeDiscoverySlug}`);
      const params = new URLSearchParams();

      if (status && status.toLowerCase() !== 'all') {
        params.set('status', status.toLowerCase());
      }

      params.set('limit', String(pageSize + 1));
      params.set('offset', String(offset));
      if (tagSlug) {
        params.set('tagSlug', tagSlug);
      }
      url.search = params.toString();

      const response = await fetch(url.toString(), { signal });

      if (!response.ok) {
        throw new Error(`Failed to fetch ${status} markets`);
      }

      const data = await response.json();
      const payload = data?.result && typeof data.result === 'object' ? data.result : data;
      const rawMarkets = Array.isArray(payload.markets) ? payload.markets : [];
      const rawPage = rawMarkets.slice(0, pageSize);
      const normalized = rawPage
        .map(normalizeMarketOverview)
        .filter((item) => item !== null);
      const visibleMarkets = groupMarketRows(normalized);
      nextOffsetRef.current = offset + rawPage.length;

      setMarketsData((current) => {
        if (!append) {
          return visibleMarkets;
        }
        const seen = new Set(current.map((item) => item.market?.marketGroup?.id
          ? `group:${item.market.marketGroup.id}`
          : `market:${item.market?.id ?? item.market?.marketId}`));
        return [
          ...current,
          ...visibleMarkets.filter((item) => {
            const key = item.market?.marketGroup?.id
              ? `group:${item.market.marketGroup.id}`
              : `market:${item.market?.id ?? item.market?.marketId}`;
            return !seen.has(key);
          }),
        ];
      });
      setHasMore(rawMarkets.length > pageSize);
    } catch (err) {
      if (err.name === 'AbortError') {
        return;
      }
      console.error(`Error fetching ${status} market data:`, err);
      setError(err.message || String(err));
    } finally {
      if (activeRequestRef.current === requestId) {
        inFlightRef.current = false;
      }
      setTimeout(() => {
        if (activeRequestRef.current === requestId) {
          setLoading(false);
          setLoadingMore(false);
        }
      }, 300);
    }
  }, [discoverySlug, pageSize, status, tagSlug]);

  useEffect(() => {
    const controller = new AbortController();
    setMarketsData([]);
    setHasMore(true);
    nextOffsetRef.current = 0;

    fetchMarketsPage({ offset: 0, append: false, signal: controller.signal });

    return () => controller.abort();
  }, [fetchMarketsPage]);

  useEffect(() => {
    if (loading || loadingMore || !hasMore || error) {
      return undefined;
    }

    if (observerRef.current) {
      observerRef.current.disconnect();
    }

    observerRef.current = new IntersectionObserver((entries) => {
      if (!entries[0]?.isIntersecting) {
        return;
      }

      const controller = new AbortController();
      fetchMarketsPage({
        offset: nextOffsetRef.current,
        append: true,
        signal: controller.signal,
      });
    }, {
      rootMargin: '300px',
    });

    const sentinel = sentinelRef.current;
    if (sentinel) {
      observerRef.current.observe(sentinel);
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [error, fetchMarketsPage, hasMore, loading, loadingMore, marketsData.length]);

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
          <div ref={sentinelRef} className="py-4 text-center text-sm text-gray-400">
            {loadingMore && (
              <div className="flex items-center justify-center gap-2">
                <LoadingSpinner />
                <span>Loading more markets...</span>
              </div>
            )}
            {!loadingMore && !hasMore && 'No more markets.'}
          </div>
        </>
      )}
    </div>
  );
}

export default MarketsByStatusTable;
