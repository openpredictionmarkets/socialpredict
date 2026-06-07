import React, { useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import PrivateUserInfoLayout from '../../components/layouts/profile/private/PrivateUserInfoLayout';
import PortfolioTabContent from '../../components/layouts/profile/public/PortfolioTabContent';
import UserFinancialStatementsLayout from '../../components/layouts/profile/public/UserFinancialStatementsLayout';
import MarketLifecycleTable from '../../components/layouts/profile/MarketLifecycleTable';
import SiteTabs from '../../components/tabs/SiteTabs';
import useUserData from '../../hooks/useUserData';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { listMyLifecycleMarkets } from '../../api/lifecycleMarketsApi';

const ErrorMessage = ({ message }) => (
  <div
    className='bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative'
    role='alert'
  >
    <strong className='font-bold'>Error:</strong>
    <span className='block sm:inline'> {message}</span>
  </div>
);

const lifecycleTabByStatus = {
  proposed: 'Proposed Markets',
  published: 'Published Markets',
  rejected: 'Rejected Markets',
};

const ProfileMarketLifecycleTab = ({ status }) => {
  const { token } = useAuth();
  const [markets, setMarkets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadMarkets = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await listMyLifecycleMarkets({ token, status });
        if (!ignore) {
          setMarkets(data.markets || []);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load market queue.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadMarkets();
    return () => {
      ignore = true;
    };
  }, [status, token]);

  if (loading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return <ErrorMessage message={error} />;
  }

  return (
    <MarketLifecycleTable
      markets={markets}
      emptyMessage={`No ${status} markets found.`}
    />
  );
};

const Profile = () => {
  const { username } = useAuth();
  const location = useLocation();
  const { userData, userLoading, userError } = useUserData(username, true);

  const defaultTab = useMemo(() => {
    const params = new URLSearchParams(location.search);
    return params.get('tab') || 'Proposed Markets';
  }, [location.search]);

  if (userLoading) {
    return <LoadingSpinner />;
  }

  if (userError) {
    return <ErrorMessage message={`Error loading user data: ${userError}`} />;
  }

  const proposedMarket = location.state?.proposedMarket;
  const marketCreationCost = location.state?.marketCreationCost;
  const isActiveModerator =
    String(userData?.usertype || '').toUpperCase() === 'MODERATOR' &&
    String(userData?.moderatorStatus || '').toLowerCase() === 'active';
  const resolvedDefaultTab = isActiveModerator ? defaultTab : 'User Info';

  const profileTabs = [
    {
      label: 'User Info',
      content: <PrivateUserInfoLayout userData={userData} />,
    },
    ...(isActiveModerator ? [
      {
        label: lifecycleTabByStatus.proposed,
        content: <ProfileMarketLifecycleTab status='proposed' />,
      },
      {
        label: lifecycleTabByStatus.published,
        content: <ProfileMarketLifecycleTab status='published' />,
      },
      {
        label: lifecycleTabByStatus.rejected,
        content: <ProfileMarketLifecycleTab status='rejected' />,
      },
    ] : []),
    {
      label: 'Portfolio',
      content: <PortfolioTabContent username={username} />,
    },
    {
      label: 'Financials',
      content: <UserFinancialStatementsLayout username={username} />,
    },
  ];

  return (
    <div className='min-h-screen bg-primary-background text-white'>
      <div className='container mx-auto px-4 py-6'>
        <div className='mb-6'>
          <p className='text-xs uppercase tracking-[0.22em] text-primary-pink'>Private Profile</p>
          <h1 className='mt-2 text-3xl font-bold'>Profile Details</h1>
          <p className='mt-2 text-gray-400'>Manage your account, proposals, published markets, rejected markets, portfolio, and financial history.</p>
        </div>

        {isActiveModerator && proposedMarket && (
          <div className='mb-6 rounded-lg border border-amber-500 bg-amber-950/40 p-4 text-amber-50'>
            <p className='text-sm uppercase tracking-[0.18em] text-amber-300'>Proposed market created</p>
            <h2 className='mt-2 text-xl font-semibold'>{proposedMarket.questionTitle}</h2>
            <p className='mt-2 text-sm'>
              Market ID <span className='font-mono'>{proposedMarket.id}</span> is awaiting admin review.
              {marketCreationCost !== undefined && marketCreationCost !== null && (
                <> The proposal cost was <span className='font-semibold'>{marketCreationCost}</span> credits.</>
              )}
            </p>
          </div>
        )}

        <SiteTabs tabs={profileTabs} defaultTab={resolvedDefaultTab} />
      </div>
    </div>
  );
};

export default Profile;
