import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { useMarketDetails } from '../../hooks/useMarketDetails';
import { useAuth } from '../../helpers/AuthContent';
import ResolveModalButton from '../../components/modals/resolution/ResolveModal';
import BetModalButton from '../../components/modals/bet/BetModal';
import ActivityTabs from '../../components/tabs/ActivityTabs';

const LoadingSpinner = () => (
  <div className='flex justify-center items-center h-64'>
    <div className='animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-blue-500'></div>
  </div>
);

const MarketDetails = () => {
  const { username } = useAuth();
  const { details, isLoggedIn, token, refetchData, currentProbability } =
    useMarketDetails();

  if (!details) {
    return <LoadingSpinner />;
  }

  const isCreator = username === details?.creator?.username;
  const isResolved = details?.market?.isResolved === true;

  return (
    <div className='flex flex-col min-h-screen p-6 space-y-8'>
      <div className='flex-grow'>
        <MarketDetailsTable
          market={details.market}
          creator={details.creator}
          numUsers={details.numUsers}
          totalVolume={details.totalVolume}
          currentProbability={currentProbability}
          probabilityChanges={details.probabilityChanges}
        />
      </div>
      <div className='flex items-center justify-start space-x-4'>
        {isCreator && !isResolved && (
          <ResolveModalButton
            marketId={details.market.id}
            token={token}
            disabled={!token}
            className='text-xs px-4 py-2'
          />
        )}
        {!isResolved && isLoggedIn && (
          <BetModalButton
            marketId={details.market.id}
            token={token}
            disabled={!token}
            onTransactionSuccess={refetchData}
            className='text-xs px-4 py-2'
          />
        )}
      </div>
      <div className='max-w-4xl mx-auto mt-8 w-full'>
        <ActivityTabs marketId={details.market.id} />
      </div>
    </div>
  );
};

export default MarketDetails;
