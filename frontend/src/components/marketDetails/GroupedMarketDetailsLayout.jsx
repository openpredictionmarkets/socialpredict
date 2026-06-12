import React, { useEffect, useMemo, useState } from 'react';
import LoadingSpinner from '../loaders/LoadingSpinner';
import GroupedMarketChart from '../charts/GroupedMarketChart';
import MarketTagChips from '../markets/MarketTagChips';
import SiteTabs from '../tabs/SiteTabs';
import TradeTabs from '../tabs/TradeTabs';
import TradeCTA from '../TradeCTA';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import { getMarketGroupDetails, getMarketSummaryReadModel } from '../../api/marketsApi';

const toNumber = (value, fallback = 0) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const uniqueTagsBySlug = (answers = []) => {
  const seen = new Set();
  const tags = [];
  answers.forEach((answer) => {
    (answer?.market?.market?.tags || []).forEach((tag) => {
      const key = tag?.slug || tag?.id || tag?.displayName;
      if (!key || seen.has(key)) {
        return;
      }
      seen.add(key);
      tags.push(tag);
    });
  });
  return tags;
};

const probabilityDisplay = (answer) => {
  const probability = toNumber(
    answer?.summary?.lastProbability
      ?? answer?.market?.lastProbability
      ?? answer?.market?.market?.initialProbability,
    0.5,
  );
  return probability.toFixed(2);
};

const canTradeMarket = (market, isLoggedIn) => {
  if (!isLoggedIn || market?.isResolved) {
    return false;
  }
  const lifecycle = String(market?.lifecycleStatus || '').toLowerCase();
  const status = String(market?.status || '').toLowerCase();
  if (lifecycle && lifecycle !== 'published') {
    return false;
  }
  if (status && status !== 'active') {
    return false;
  }
  const closeDate = market?.resolutionDateTime ? new Date(market.resolutionDateTime) : null;
  if (!closeDate || Number.isNaN(closeDate.getTime())) {
    return false;
  }
  return closeDate > new Date();
};

export default function GroupedMarketDetailsLayout({
  marketGroup,
  fallbackMarket,
  creator,
  isLoggedIn,
  token,
  refetchData,
}) {
  const [groupData, setGroupData] = useState(null);
  const [answers, setAnswers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showFullDescription, setShowFullDescription] = useState(false);
  const [showTradeModal, setShowTradeModal] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let ignore = false;

    const loadGroup = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await getMarketGroupDetails(marketGroup.id);
        const rawAnswers = [...(data?.answers || [])].sort((left, right) => (
          Number(left.displayOrder || 0) - Number(right.displayOrder || 0)
        ));
        const summaries = await Promise.all(rawAnswers.map((answer) => (
          getMarketSummaryReadModel(answer.marketId).catch(() => null)
        )));
        if (!ignore) {
          setGroupData(data);
          setAnswers(rawAnswers.map((answer, index) => ({
            ...answer,
            summary: summaries[index],
          })));
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Failed to load grouped market.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    if (marketGroup?.id) {
      loadGroup();
    }

    return () => {
      ignore = true;
    };
  }, [marketGroup?.id, refreshKey]);

  const group = groupData?.group || marketGroup || {};
  const groupCreator = groupData?.creator || creator || {};
  const tags = useMemo(() => uniqueTagsBySlug(answers), [answers]);
  const aggregate = useMemo(() => ({
    users: Math.max(0, ...answers.map((answer) => toNumber(answer?.summary?.numUsers ?? answer?.market?.numUsers))),
    volume: answers.reduce((sum, answer) => sum + toNumber(answer?.summary?.totalVolume ?? answer?.market?.totalVolume), 0),
    dust: answers.reduce((sum, answer) => sum + toNumber(answer?.summary?.marketDust ?? answer?.market?.marketDust), 0),
  }), [answers]);

  const closeDate = fallbackMarket?.resolutionDateTime || group?.resolutionDateTime;
  const sortedAnswers = useMemo(() => [...answers].sort((left, right) => (
    Number(left.displayOrder || 0) - Number(right.displayOrder || 0)
  )), [answers]);
  const anyTradableAnswer = sortedAnswers.some((answer) => canTradeMarket(answer?.market?.market || {}, isLoggedIn));
  const tradeButtonLabel = (() => {
    if (!isLoggedIn) {
      return 'LOG IN TO TRADE';
    }
    if (anyTradableAnswer) {
      return 'TRADE';
    }
    const hasProposed = sortedAnswers.some((answer) => (
      String(answer?.market?.market?.lifecycleStatus || '').toLowerCase() === 'proposed'
    ));
    return hasProposed ? 'AWAITING APPROVAL' : 'TRADING CLOSED';
  })();

  const handleTransactionSuccess = () => {
    setShowTradeModal(false);
    setRefreshKey((current) => current + 1);
    if (refetchData) {
      refetchData();
    }
  };

  const answerTradeTabs = sortedAnswers.map((answer) => {
    const childMarket = answer?.market?.market || {};
    const tradable = canTradeMarket(childMarket, isLoggedIn);
    return {
      label: answer.answerLabel || `Answer ${answer.displayOrder + 1}`,
      content: tradable ? (
        <TradeTabs
          marketId={answer.marketId}
          market={childMarket}
          token={token}
          onTransactionSuccess={handleTransactionSuccess}
        />
      ) : (
        <div className='rounded-lg bg-blue-950/70 p-4 text-sm text-blue-50'>
          This answer is not open for trading.
        </div>
      ),
    };
  });

  if (loading) {
    return (
      <div className='rounded-lg bg-gray-900 p-6 text-gray-300'>
        <LoadingSpinner />
        Loading grouped market...
      </div>
    );
  }

  if (error) {
    return (
      <div className='rounded-lg border border-red-500 bg-red-950/50 p-4 text-red-100'>
        {error}
      </div>
    );
  }

  return (
    <div className='bg-gray-900 text-gray-300 p-4 rounded-lg shadow-lg w-full'>
      <section className='mb-4'>
        <h1 className='text-xl font-semibold text-white mb-2 break-words line-clamp-2'>
          {group.questionTitle || marketGroup.questionTitle || fallbackMarket.questionTitle}
        </h1>
        <div className='flex flex-wrap items-center gap-2 text-sm text-gray-400'>
          <span>@{groupCreator.username || group.creatorUsername || fallbackMarket.creatorUsername}</span>
          <span>•</span>
          <span>{answers.length} answers</span>
          <span>•</span>
          <span>Closes {formatResolutionDate(closeDate)}</span>
        </div>
        <MarketTagChips tags={tags} className='mt-3' />
      </section>

      <div className='mb-4'>
        <GroupedMarketChart answers={answers} title='Probability Changes' />
      </div>

      {group.description && (
        <>
          <div className='mb-4'>
            <button
              type='button'
              onClick={() => setShowFullDescription(!showFullDescription)}
              className='w-full py-2 bg-gray-700 hover:bg-gray-600 transition-colors duration-200 rounded-lg text-center text-sm'
            >
              {showFullDescription ? 'Hide Contract Text' : 'Show Full Contract Text'}
            </button>
          </div>
          {showFullDescription && (
            <div className='mb-4 rounded-lg bg-gray-800 p-4 text-sm'>
              <p className='whitespace-pre-wrap'>{group.description}</p>
            </div>
          )}
        </>
      )}

      <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
        {[
          { label: 'Answers', value: answers.length, icon: '▦' },
          { label: 'Users', value: aggregate.users, icon: '👤' },
          { label: 'Volume', value: Math.round(aggregate.volume), icon: '📊' },
          { label: 'Closes', value: formatResolutionDate(closeDate), icon: '📅' },
        ].map((item) => (
          <div key={item.label} className='rounded-lg bg-gray-800 p-2'>
            <div className='text-xs text-gray-400'>{item.label}</div>
            <div className='truncate text-sm font-semibold'>
              {item.icon} {item.value}
            </div>
          </div>
        ))}
      </div>

      {aggregate.dust > 0 && (
        <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
          <div className='bg-gray-800 p-2 rounded-lg'>
            <div className='text-xs text-gray-400'>Dust</div>
            <div className='text-sm font-semibold truncate'>✨ {aggregate.dust}</div>
          </div>
        </div>
      )}

      <div className='mb-4 grid gap-2 sm:grid-cols-3'>
        {sortedAnswers.map((answer) => (
          <div key={answer.id || answer.marketId} className='rounded-lg bg-gray-800 p-3 text-center'>
            <div className='truncate text-sm font-semibold text-white'>{answer.answerLabel}</div>
            <div className='mt-1 text-xs text-gray-400'>YES {probabilityDisplay(answer)}</div>
          </div>
        ))}
      </div>

      <div className='flex items-center justify-center mb-4 space-x-4 py-4'>
        <button
          type='button'
          disabled={!anyTradableAnswer}
          onClick={() => setShowTradeModal(true)}
          className='min-w-32 rounded border bg-custom-gray-light px-4 py-2 text-xs text-white transition hover:bg-neutral-btn disabled:cursor-not-allowed disabled:opacity-50 sm:text-sm md:text-base'
        >
          {tradeButtonLabel}
        </button>
      </div>

      {anyTradableAnswer && (
        <TradeCTA onClick={() => setShowTradeModal(true)} disabled={!token} />
      )}

      <div className='h-32 md:hidden' />

      {showTradeModal && (
        <div className='fixed inset-0 z-50 flex items-center justify-center bg-gray-600 bg-opacity-50'>
          <div className='bet-modal relative m-6 mx-auto rounded-lg bg-blue-900 p-6 text-white' style={{ width: '350px' }}>
            <SiteTabs tabs={answerTradeTabs} />
            <button
              type='button'
              onClick={() => setShowTradeModal(false)}
              className='absolute right-0 top-0 mr-4 mt-4 text-gray-400 hover:text-white'
            >
              x
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
