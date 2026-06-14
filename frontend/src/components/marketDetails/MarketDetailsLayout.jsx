import React, { useState } from 'react';
import ResolutionAlert from '../resolutions/ResolutionAlert';
import MarketChart from '../charts/MarketChart';
import ActivityTabs from '../../components/tabs/ActivityTabs';
import ResolveModalButton from '../modals/resolution/ResolveModal';
import TradeCTA from '../TradeCTA';
import TradeTabs from '../../components/tabs/TradeTabs';
import { BetButton } from '../buttons/trade/BetButtons';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import StewardTag, { stewardUsernameFor } from '../markets/StewardTag';
import MarketTagChips from '../markets/MarketTagChips';
import MarkdownLite from '../markdown/MarkdownLite';
import GroupedMarketDetailsLayout from './GroupedMarketDetailsLayout';
import { proposeMarketDescriptionAmendment } from '../../api/marketDescriptionAmendmentsApi';

const DEFAULT_CREATOR_EMOJI = '👤';

function MarketDetailsTable({
  market,
  creator,
  numUsers,
  totalVolume,
  marketDust,
  currentProbability,
  probabilityChanges,
  freshness,
  descriptionAmendments = [],
  marketId: marketIdProp,
  username,
  usertype,
  moderatorStatus,
  isLoggedIn,
  token,
  refetchData,
}) {
  const safeMarket = market ?? {};
  const safeCreator = creator ?? {};
  const resolvedMarketId = marketIdProp ?? safeMarket.id;
  const creatorUsername = safeMarket.creatorUsername ?? safeCreator.username ?? 'unknown';
  const stewardUsername = stewardUsernameFor(safeMarket, creatorUsername);
  const creatorEmoji = safeCreator.personalEmoji ?? DEFAULT_CREATOR_EMOJI;
  const marketDescription = safeMarket.description ?? '';
  const freshnessGeneratedAt = freshness?.generatedAt
    ? new Date(freshness.generatedAt)
    : null;
  const freshnessLabel = freshnessGeneratedAt && !Number.isNaN(freshnessGeneratedAt.getTime())
    ? freshnessGeneratedAt.toLocaleTimeString()
    : '';

  const [showFullDescription, setShowFullDescription] = useState(false);
  const [showBetModal, setShowBetModal] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [amendmentBody, setAmendmentBody] = useState('');
  const [amendmentReason, setAmendmentReason] = useState('');
  const [amendmentMessage, setAmendmentMessage] = useState('');
  const [amendmentError, setAmendmentError] = useState('');
  const [submittingAmendment, setSubmittingAmendment] = useState(false);

  const toggleBetModal = () => setShowBetModal(prev => !prev);

  const handleMarketResolved = () => {
    if (refetchData) {
      refetchData();
    }
    setRefreshTrigger(prev => prev + 1);
  };

  const handleTransactionSuccess = () => {
    setShowBetModal(false);  // Close modal
    if (refetchData) {
      refetchData();  // Trigger data refresh
    }
    setRefreshTrigger(prev => prev + 1); // Trigger positions refresh
  };

  const shouldShowTradeButtons =
    !safeMarket.isResolved &&
    isLoggedIn &&
    safeMarket.resolutionDateTime &&
    new Date(safeMarket.resolutionDateTime) > new Date();
  const canResolveMarket =
    !safeMarket.isResolved &&
    String(username || '').trim() === String(stewardUsername || '').trim();
  const canProposeDescriptionAmendment =
    isLoggedIn &&
    token &&
    !safeMarket.isResolved &&
    ['proposed', 'published', 'active', ''].includes(String(safeMarket.lifecycleStatus || safeMarket.status || '').toLowerCase()) &&
    String(username || '').trim() === String(stewardUsername || '').trim();

  const submitDescriptionAmendment = async (event) => {
    event.preventDefault();
    setAmendmentMessage('');
    setAmendmentError('');
    setSubmittingAmendment(true);
    try {
      await proposeMarketDescriptionAmendment({
        token,
        marketId: resolvedMarketId,
        body: amendmentBody,
        submitReason: amendmentReason,
      });
      setAmendmentBody('');
      setAmendmentReason('');
      setAmendmentMessage('Description amendment submitted for admin review.');
    } catch (err) {
      setAmendmentError(err.message || 'Unable to submit description amendment.');
    } finally {
      setSubmittingAmendment(false);
    }
  };

  if (safeMarket.marketGroup?.id) {
    return (
      <GroupedMarketDetailsLayout
        marketGroup={safeMarket.marketGroup}
        fallbackMarket={safeMarket}
        creator={safeCreator}
        isLoggedIn={isLoggedIn}
        token={token}
        username={username}
        usertype={usertype}
        moderatorStatus={moderatorStatus}
        refetchData={refetchData}
      />
    );
  }

  return (
    <div className='bg-gray-900 text-gray-300 p-4 rounded-lg shadow-lg w-full'>
      <ResolutionAlert
        isResolved={safeMarket.isResolved}
        resolutionResult={safeMarket.resolutionResult}
        market={safeMarket}
      />

      <div className='mb-4'>
        <h1
          className='text-xl font-semibold text-white mb-2 break-words line-clamp-2'
          title={safeMarket.questionTitle}
        >
          {safeMarket.questionTitle}
        </h1>
        <div className='flex flex-wrap items-center gap-2 text-sm text-gray-400'>
          <a
            href={`/user/${creatorUsername}`}
            className='hover:text-blue-400 transition-colors duration-200'
          >
            <span role='img' aria-label='Creator'>
              {creatorEmoji}
            </span>
            @{creatorUsername}
          </a>
          <StewardTag username={stewardUsername} creatorUsername={creatorUsername} />
          <span>•</span>
          <span>🪙 {currentProbability.toFixed(2)}</span>
        </div>
        <MarketTagChips tags={safeMarket.tags || []} className='mt-3' />
      </div>

      <div className='mb-4'>
        <MarketChart
          data={probabilityChanges}
          currentProbability={currentProbability}
          title='Probability Changes'
          className='w-full'
          closeDateTime={safeMarket.resolutionDateTime}
          yesLabel={safeMarket.yesLabel}
          noLabel={safeMarket.noLabel}
        />
      </div>

      <div className='mb-4'>
        <button
          onClick={() => setShowFullDescription(!showFullDescription)}
          className='w-full py-2 bg-gray-700 hover:bg-gray-600 transition-colors duration-200 rounded-lg text-center text-sm'
        >
          {showFullDescription ? 'Hide Contract Text' : 'Show Full Contract Text'}
        </button>
      </div>
      {showFullDescription && (
        <div className='mb-4 bg-gray-800 p-4 rounded-lg'>
          <div
            className="grid gap-4 text-sm break-words"
            style={{
              wordBreak: 'break-word',
              overflowWrap: 'break-word',
              hyphens: 'auto',
            }}
          >
            <section>
              <h2 className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">
                Description
              </h2>
              {marketDescription.trim() ? (
                <p className="whitespace-pre-wrap">{marketDescription}</p>
              ) : (
                <p className="text-gray-500 italic">No description provided.</p>
              )}
            </section>
            {descriptionAmendments.length > 0 && (
              <section className="grid gap-3">
                <h2 className="text-sm font-semibold uppercase tracking-[0.14em] text-sky-200">Amendments</h2>
                {descriptionAmendments.map((amendment, index) => (
                  <article key={amendment.id || amendment.version} className="rounded-md border border-sky-900/70 bg-sky-950/30 p-3">
                    <div className="mb-2 flex flex-wrap gap-2 text-xs text-sky-100/80">
                      <span>Amendment {index + 1}</span>
                      <span>Submitted by @{amendment.createdBy}</span>
                      {amendment.approvedAt && <span>Approved {new Date(amendment.approvedAt).toLocaleString()}</span>}
                    </div>
                    <MarkdownLite className="text-gray-200">{amendment.body}</MarkdownLite>
                  </article>
                ))}
              </section>
            )}
          </div>
        </div>
      )}

      {canProposeDescriptionAmendment && (
        <form onSubmit={submitDescriptionAmendment} className="mb-4 grid gap-3 rounded-lg border border-sky-800/60 bg-sky-950/20 p-4">
          <div>
            <p className="text-sm font-semibold text-sky-100">Propose Description Amendment</p>
            <p className="mt-1 text-xs text-sky-100/70">
              Titles are immutable. Description changes are append-only, versioned, and require admin approval before becoming public contract text.
            </p>
          </div>
          {amendmentMessage && (
            <div className="rounded-md bg-emerald-700 p-3 text-sm text-white">{amendmentMessage}</div>
          )}
          {amendmentError && (
            <div className="rounded-md bg-red-700 p-3 text-sm text-white">{amendmentError}</div>
          )}
          <textarea
            value={amendmentBody}
            onChange={(event) => setAmendmentBody(event.target.value)}
            rows={5}
            maxLength={2000}
            placeholder="Append clarification using markdown-lite: bold, italic, inline code, links, lists, and blockquotes. Raw HTML is not allowed."
            className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
            required
          />
          {amendmentBody && (
            <div className="rounded-md border border-gray-700 bg-gray-950 p-3">
              <p className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">Preview</p>
              <MarkdownLite>{amendmentBody}</MarkdownLite>
            </div>
          )}
          <input
            type="text"
            value={amendmentReason}
            onChange={(event) => setAmendmentReason(event.target.value)}
            maxLength={500}
            placeholder="Optional submit reason for admin review"
            className="rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
          <button
            type="submit"
            disabled={submittingAmendment || !amendmentBody.trim()}
            className="justify-self-start rounded-md bg-sky-700 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submittingAmendment ? 'Submitting...' : 'Submit Amendment for Review'}
          </button>
        </form>
      )}

      <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
        {[
          { label: 'Users', value: `${numUsers}`, icon: '👤' },
          { 
            label: 'Volume', 
            value: `${Math.round(totalVolume)}`,
            icon: '📊' 
          },
          { label: 'Comments', value: '0', icon: '💬' },
          {
            label: 'Closes',
            value: safeMarket.isResolved
              ? 'Closed'
              : formatResolutionDate(safeMarket.resolutionDateTime),
            icon: '📅',
          },
        ].map((item, index) => (
          <div key={index} className='bg-gray-800 p-2 rounded-lg'>
            <div className='text-xs text-gray-400'>{item.label}</div>
            <div className='text-sm font-semibold truncate'>
              {item.icon} {item.value}
            </div>
          </div>
        ))}
      </div>

      {marketDust > 0 && (
        <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center mb-4'>
          <div className='bg-gray-800 p-2 rounded-lg'>
            <div className='text-xs text-gray-400'>Dust</div>
            <div className='text-sm font-semibold truncate'>
              ✨ {marketDust}
            </div>
          </div>
          <div className='bg-gray-800 p-2 rounded-lg opacity-50'>
            <div className='text-xs text-gray-400'>—</div>
            <div className='text-sm font-semibold truncate'>—</div>
          </div>
          <div className='bg-gray-800 p-2 rounded-lg opacity-50'>
            <div className='text-xs text-gray-400'>—</div>
            <div className='text-sm font-semibold truncate'>—</div>
          </div>
          <div className='bg-gray-800 p-2 rounded-lg opacity-50'>
            <div className='text-xs text-gray-400'>—</div>
            <div className='text-sm font-semibold truncate'>—</div>
          </div>
        </div>
      )}

      {freshnessLabel && (
        <p className="mb-4 text-center text-xs text-gray-500">
          Display widgets generated at {freshnessLabel}. Trade confirmations remain authoritative.
        </p>
      )}

      <div className='flex items-center justify-center mb-4 space-x-4 py-4'>
        {canResolveMarket && (
          <ResolveModalButton
            marketId={resolvedMarketId}
            token={token}
            market={market}
            onResolved={handleMarketResolved}
            disabled={!token}
            className='text-xs px-4 py-2'
          />
        )}
        {shouldShowTradeButtons && (
          <div className="hidden md:block">
            <BetButton onClick={toggleBetModal} className="text-xs px-4 py-2" />
          </div>
        )}
      </div>

      <div className='mx-auto w-full mb-4'>
        <ActivityTabs marketId={resolvedMarketId} market={safeMarket} refreshTrigger={refreshTrigger} />
      </div>

      {/* Mobile floating CTA */}
      {shouldShowTradeButtons && (
        <TradeCTA onClick={toggleBetModal} disabled={!token} />
      )}

      {/* Spacer so content doesn't sit under the CTA */}
      <div className="h-32 md:hidden" />

      {/* Shared Trade Modal */}
      {showBetModal && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bet-modal relative bg-blue-900 p-6 rounded-lg text-white m-6 mx-auto" style={{ width: '350px' }}>
            <TradeTabs
              marketId={resolvedMarketId}
              market={safeMarket}
              token={token}
              onTransactionSuccess={handleTransactionSuccess}
            />
            <button onClick={toggleBetModal} className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white">
              ✕
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default MarketDetailsTable;
