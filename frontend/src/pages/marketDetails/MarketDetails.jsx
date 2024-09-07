import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { useMarketDetails } from '../../hooks/useMarketDetails';
import { useAuth } from '../../helpers/AuthContent';
import ResolveModalButton from '../../components/modals/resolution/ResolveModal';
import BetModalButton from '../../components/modals/bet/BetModal';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';

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
    <div className='flex flex-col space-y-8'>
      <div className='flex-grow'>
        <MarketDetailsTable
          market={details.market}
          creator={details.creator}
          numUsers={details.numUsers}
          totalVolume={details.totalVolume}
          currentProbability={currentProbability}
          probabilityChanges={details.probabilityChanges}
          marketId={details.market.id}
        />
      </div>
      <div className='flex items-center justify-center  space-x-4'>
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
    </div>
  );
};

export default MarketDetails;
