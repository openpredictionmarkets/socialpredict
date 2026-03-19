import React from 'react';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { useMarketDetails } from '../../hooks/useMarketDetails';
import { useAuth } from '../../helpers/AuthContent';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { useDocumentMeta } from '../../hooks/useDocumentMeta';

const MarketDetails = () => {
  const { username } = useAuth();
  const { details, isLoggedIn, token, refetchData, currentProbability } =
    useMarketDetails();

  useDocumentMeta({
    title: details ? `${details.market.question} — SocialPredict` : 'SocialPredict',
    description: details
      ? `${Math.round(currentProbability * 100)}% probability · created by ${details.creator}`
      : undefined,
  });

  if (!details) {
    return <LoadingSpinner />;
  }

  return (
    <div className='flex flex-col min-h-screen'>
      <div className='flex-grow overflow-y-auto'>
        <MarketDetailsTable
          market={details.market}
          creator={details.creator}
          numUsers={details.numUsers}
          totalVolume={details.totalVolume}
          marketDust={details.marketDust || 0}
          currentProbability={currentProbability}
          probabilityChanges={details.probabilityChanges}
          marketId={details.market.id}
          username={username}
          isLoggedIn={isLoggedIn}
          token={token}
          refetchData={refetchData}
        />
      </div>
    </div>
  );
};

export default MarketDetails;
