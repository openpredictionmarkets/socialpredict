import React, { useEffect, useMemo, useState } from 'react';
import LoadingSpinner from '../loaders/LoadingSpinner';
import GroupedMarketChart from '../charts/GroupedMarketChart';
import MarketTagChips from '../markets/MarketTagChips';
import TradeTabs from '../tabs/TradeTabs';
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

const childMarketStatus = (market) => (
  market?.lifecycleStatus || market?.status || 'unknown'
);

const canTradeMarket = (market, isLoggedIn) => {
  if (!isLoggedIn || market?.isResolved) {
    return false;
  }
  const closeDate = market?.resolutionDateTime ? new Date(market.resolutionDateTime) : null;
  if (!closeDate || Number.isNaN(closeDate.getTime())) {
    return false;
  }
  return closeDate > new Date();
};

function AnswerTradeCard({ answer, isLoggedIn, onTrade }) {
  const childMarket = answer?.market?.market || {};
  const users = toNumber(answer?.summary?.numUsers ?? answer?.market?.numUsers);
  const volume = toNumber(answer?.summary?.totalVolume ?? answer?.market?.totalVolume);
  const marketDust = toNumber(answer?.summary?.marketDust ?? answer?.market?.marketDust);
  const answerLabel = answer?.answerLabel || childMarket.marketGroup?.answerLabel || childMarket.questionTitle || 'Answer';
  const tradable = canTradeMarket(childMarket, isLoggedIn);

  return (
    <article className='rounded-xl border border-gray-700 bg-gray-900/80 p-4 shadow-lg'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-xs uppercase tracking-[0.18em] text-primary-pink'>Answer</p>
          <h2 className='mt-1 text-xl font-bold text-white'>{answerLabel}</h2>
          <p className='mt-1 text-xs text-gray-500'>Child market #{answer.marketId}</p>
        </div>
        <span className='rounded-full border border-gray-600 bg-gray-800 px-3 py-1 text-xs font-semibold uppercase text-gray-200'>
          {childMarketStatus(childMarket)}
        </span>
      </div>

      <div className='mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4'>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>YES</p>
          <p className='mt-1 text-lg font-bold text-white'>{probabilityDisplay(answer)}</p>
        </div>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Users</p>
          <p className='mt-1 text-lg font-bold text-white'>{users}</p>
        </div>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Volume</p>
          <p className='mt-1 text-lg font-bold text-white'>{volume}</p>
        </div>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Dust</p>
          <p className='mt-1 text-lg font-bold text-white'>{marketDust}</p>
        </div>
      </div>

      <div className='mt-5 flex flex-wrap items-center justify-between gap-3'>
        <p className='text-sm text-gray-400'>
          Trade YES/NO on this answer independently.
        </p>
        <button
          type='button'
          disabled={!tradable}
          onClick={() => onTrade(answer)}
          className='rounded-lg bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-pink/80 disabled:cursor-not-allowed disabled:bg-gray-700 disabled:text-gray-400'
        >
          {tradable ? 'Trade This Answer' : 'Trading Closed'}
        </button>
      </div>
    </article>
  );
}

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
  const [selectedAnswer, setSelectedAnswer] = useState(null);
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
  const selectedMarket = selectedAnswer?.market?.market || {};

  const handleTransactionSuccess = () => {
    setSelectedAnswer(null);
    setRefreshKey((current) => current + 1);
    if (refetchData) {
      refetchData();
    }
  };

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
        <p className='text-xs uppercase tracking-[0.2em] text-primary-pink'>
          Multiple-Choice Binary Market
        </p>
        <h1 className='mt-2 text-2xl font-bold text-white sm:text-3xl'>
          {group.questionTitle || marketGroup.questionTitle || fallbackMarket.questionTitle}
        </h1>
        <div className='mt-3 flex flex-wrap items-center gap-2 text-sm text-gray-400'>
          <span>@{groupCreator.username || group.creatorUsername || fallbackMarket.creatorUsername}</span>
          <span>•</span>
          <span>{answers.length} answers</span>
          <span>•</span>
          <span>Closes {formatResolutionDate(closeDate)}</span>
        </div>
        {(group.description || fallbackMarket.description) && (
          <p className='mt-3 max-w-4xl whitespace-pre-wrap text-sm leading-6 text-gray-300'>
            {group.description || fallbackMarket.description}
          </p>
        )}
        <MarketTagChips tags={tags} className='mt-3' />
        <div className='mt-4 rounded-xl border border-cyan-500/30 bg-cyan-950/30 p-4 text-sm text-cyan-50'>
          Each answer below is a separate YES/NO market. The chart compares each answer's YES probability; probabilities are independent and are not forced to add up to 100%.
        </div>
      </section>

      <div className='mb-4 grid grid-cols-2 gap-2 text-center sm:grid-cols-4'>
        {[
          { label: 'Answers', value: answers.length, icon: '▦' },
          { label: 'Users', value: aggregate.users, icon: '👤' },
          { label: 'Volume', value: Math.round(aggregate.volume), icon: '📊' },
          { label: 'Dust', value: aggregate.dust, icon: '✨' },
        ].map((item) => (
          <div key={item.label} className='rounded-lg bg-gray-800 p-2'>
            <div className='text-xs text-gray-400'>{item.label}</div>
            <div className='truncate text-sm font-semibold'>
              {item.icon} {item.value}
            </div>
          </div>
        ))}
      </div>

      <div className='mb-5'>
        <GroupedMarketChart answers={answers} title='Answer Probability Comparison' />
      </div>

      <section className='grid gap-4'>
        {answers.map((answer) => (
          <AnswerTradeCard
            key={answer.id || answer.marketId}
            answer={answer}
            isLoggedIn={isLoggedIn}
            onTrade={setSelectedAnswer}
          />
        ))}
      </section>

      {selectedAnswer && (
        <div className='fixed inset-0 z-50 flex items-center justify-center bg-gray-600 bg-opacity-50'>
          <div className='bet-modal relative m-6 mx-auto rounded-lg bg-blue-900 p-6 text-white' style={{ width: '350px' }}>
            <div className='mb-3 rounded-md bg-blue-950/70 p-3 text-sm text-blue-50'>
              Trading answer: <span className='font-semibold'>{selectedAnswer.answerLabel}</span>
            </div>
            <TradeTabs
              marketId={selectedAnswer.marketId}
              market={selectedMarket}
              token={token}
              onTransactionSuccess={handleTransactionSuccess}
            />
            <button
              type='button'
              onClick={() => setSelectedAnswer(null)}
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
