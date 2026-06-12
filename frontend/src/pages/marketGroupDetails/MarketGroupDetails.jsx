import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import MarketTagChips from '../../components/markets/MarketTagChips';
import { getMarketGroupDetails } from '../../api/marketsApi';

const formatPercent = (value) => {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return '0.50%';
  }
  return `${numeric.toFixed(2)}%`;
};

const statusLabel = (market) => (
  market?.lifecycleStatus || market?.status || 'unknown'
);

const MarketGroupAnswerCard = ({ answer }) => {
  const overview = answer?.market || {};
  const market = overview.market || {};
  const tags = market.tags || [];
  const probability = overview.lastProbability ?? market.initialProbability ?? 0.5;

  return (
    <article className='rounded-xl border border-gray-700 bg-gray-900/70 p-4 shadow-lg'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-xs uppercase tracking-[0.18em] text-primary-pink'>
            Answer Option
          </p>
          <h2 className='mt-1 text-xl font-bold text-white'>
            {answer.answerLabel}
          </h2>
          <p className='mt-2 text-sm text-gray-400'>
            Child Market #{answer.marketId}
          </p>
        </div>
        <span className='rounded-full border border-gray-600 bg-gray-800 px-3 py-1 text-xs font-semibold uppercase text-gray-200'>
          {statusLabel(market)}
        </span>
      </div>

      <div className='mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3'>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Probability</p>
          <p className='mt-1 text-lg font-bold text-white'>{formatPercent(probability)}</p>
        </div>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Users</p>
          <p className='mt-1 text-lg font-bold text-white'>{overview.numUsers ?? 0}</p>
        </div>
        <div className='rounded-lg bg-gray-800/80 p-3'>
          <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Volume</p>
          <p className='mt-1 text-lg font-bold text-white'>{overview.totalVolume ?? 0}</p>
        </div>
      </div>

      <MarketTagChips tags={tags} className='mt-4' />

      <div className='mt-5 flex flex-wrap items-center justify-between gap-3'>
        <p className='text-sm text-gray-400'>
          Trade this answer by opening its normal binary market.
        </p>
        <Link
          to={`/markets/${answer.marketId}`}
          className='rounded-lg bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-pink/80'
        >
          Open Market
        </Link>
      </div>
    </article>
  );
};

function MarketGroupDetails() {
  const { groupId } = useParams();
  const [groupData, setGroupData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadGroup = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await getMarketGroupDetails(groupId);
        if (!ignore) {
          setGroupData(data);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Failed to load market group.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadGroup();
    return () => {
      ignore = true;
    };
  }, [groupId]);

  if (loading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className='mx-auto max-w-4xl p-6'>
        <div className='rounded-lg border border-red-500 bg-red-950/50 p-4 text-red-100'>
          {error}
        </div>
      </div>
    );
  }

  const group = groupData?.group || {};
  const answers = [...(groupData?.answers || [])].sort((left, right) => left.displayOrder - right.displayOrder);

  return (
    <main className='mx-auto max-w-5xl p-4 sm:p-6'>
      <section className='rounded-2xl border border-gray-700 bg-gray-900/70 p-5 shadow-xl'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div>
            <p className='text-xs uppercase tracking-[0.2em] text-primary-pink'>
              Multiple-Choice Binary Group
            </p>
            <h1 className='mt-2 text-2xl font-bold text-white sm:text-3xl'>
              {group.questionTitle}
            </h1>
            {group.description && (
              <p className='mt-3 max-w-3xl whitespace-pre-wrap text-sm leading-6 text-gray-300'>
                {group.description}
              </p>
            )}
          </div>
          <span className='rounded-full border border-gray-600 bg-gray-800 px-3 py-1 text-xs font-semibold uppercase text-gray-200'>
            {group.lifecycleStatus || group.status || 'unknown'}
          </span>
        </div>

        <div className='mt-5 rounded-xl border border-cyan-500/30 bg-cyan-950/30 p-4 text-sm text-cyan-50'>
          Each answer below is a separate YES/NO child market. Probabilities are independent and are not normalized to add up to 100%.
        </div>

        <div className='mt-5 grid grid-cols-1 gap-3 text-sm text-gray-300 sm:grid-cols-3'>
          <div className='rounded-lg bg-gray-800/80 p-3'>
            <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Answers</p>
            <p className='mt-1 font-bold text-white'>{group.answerCount || answers.length}</p>
          </div>
          <div className='rounded-lg bg-gray-800/80 p-3'>
            <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Proposal Cost</p>
            <p className='mt-1 font-bold text-white'>{group.proposalCost ?? 0} credits</p>
          </div>
          <div className='rounded-lg bg-gray-800/80 p-3'>
            <p className='text-xs uppercase tracking-[0.14em] text-gray-400'>Creator</p>
            <p className='mt-1 font-bold text-white'>@{group.creatorUsername}</p>
          </div>
        </div>
      </section>

      <section className='mt-6 space-y-4'>
        {answers.length > 0 ? (
          answers.map((answer) => (
            <MarketGroupAnswerCard key={answer.id || answer.marketId} answer={answer} />
          ))
        ) : (
          <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-4 text-gray-300'>
            No child markets are linked to this group yet.
          </div>
        )}
      </section>
    </main>
  );
}

export default MarketGroupDetails;
