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
import { listMyMarketDescriptionAmendments } from '../../api/marketDescriptionAmendmentsApi';
import MarkdownLite from '../../components/markdown/MarkdownLite';

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

const amendmentTabByStatus = {
  pending: 'Pending Amendments',
  approved: 'Approved Amendments',
  rejected: 'Rejected Amendments',
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

const ProfileDescriptionAmendmentTab = ({ status }) => {
  const { token } = useAuth();
  const [amendments, setAmendments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;

    const loadAmendments = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await listMyMarketDescriptionAmendments({ token, status, limit: 100 });
        if (!ignore) {
          setAmendments(data.amendments || []);
        }
      } catch (err) {
        if (!ignore) {
          setError(err.message || 'Unable to load description amendments.');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    loadAmendments();
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

  if (amendments.length === 0) {
    return (
      <div className='rounded-lg border border-gray-700 bg-gray-900/70 p-6 text-center text-gray-300'>
        No {status} description amendments found.
      </div>
    );
  }

  return (
    <div className='grid gap-4'>
      {amendments.map((amendment) => {
        const previousAmendments = Array.isArray(amendment.previousApprovedAmendments)
          ? amendment.previousApprovedAmendments
          : [];
        return (
          <article key={amendment.id} className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
            <div className='flex flex-wrap items-center gap-2 text-sm text-gray-300'>
              <span className='rounded-full border border-sky-500/40 bg-sky-950/50 px-2 py-0.5 text-xs font-semibold text-sky-100'>
                Market #{amendment.marketId}
              </span>
              <span className='rounded-full border border-gray-600 bg-gray-800 px-2 py-0.5 text-xs font-semibold uppercase tracking-[0.14em] text-gray-200'>
                Amendment {Math.max(1, Number(amendment.version || 2) - 1)}
              </span>
              <span>{amendment.createdAt ? new Date(amendment.createdAt).toLocaleString() : ''}</span>
            </div>
            <a
              href={`/markets/${amendment.marketId}`}
              className='text-lg font-semibold text-white underline decoration-sky-500/40 underline-offset-4 transition hover:text-sky-200'
            >
              {amendment.marketTitle || `Market #${amendment.marketId}`}
            </a>
            <div className='grid gap-3 rounded-md border border-gray-700 bg-gray-950 p-4'>
              <div>
                <p className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-gray-400'>Description</p>
                <p className='whitespace-pre-wrap text-sm leading-6 text-gray-200'>
                  {amendment.marketDescription || 'No market description was returned.'}
                </p>
              </div>
              {previousAmendments.length > 0 && (
                <div className='grid gap-2'>
                  {previousAmendments.map((previous) => (
                    <article key={previous.id || previous.version} className='rounded-md border border-gray-700 bg-gray-900 p-3'>
                      <div className='mb-2 flex flex-wrap gap-2 text-xs text-gray-400'>
                        <span>Amendment {Math.max(1, Number(previous.version || 2) - 1)}</span>
                        <span>Approved by @{previous.approvedBy || 'admin'}</span>
                        {previous.approvedAt && <span>{new Date(previous.approvedAt).toLocaleString()}</span>}
                      </div>
                      <MarkdownLite>{previous.body}</MarkdownLite>
                    </article>
                  ))}
                </div>
              )}
            </div>
            {amendment.submitReason && (
              <div className='rounded-md border border-gray-700 bg-gray-950 p-3 text-sm text-gray-300'>
                <span className='font-semibold text-gray-100'>Submit reason:</span> {amendment.submitReason}
              </div>
            )}
            <div className='rounded-md border border-sky-900/70 bg-sky-950/20 p-4'>
              <p className='mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-sky-200'>
                Proposed Amendment {Math.max(1, Number(amendment.version || 2) - 1)}
              </p>
              <MarkdownLite>{amendment.body}</MarkdownLite>
            </div>
            {status === 'rejected' && amendment.rejectionReason && (
              <div className='rounded-md border border-rose-800/70 bg-rose-950/30 p-3 text-sm text-rose-100'>
                Rejection reason: {amendment.rejectionReason}
              </div>
            )}
          </article>
        );
      })}
    </div>
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
      {
        label: amendmentTabByStatus.pending,
        content: <ProfileDescriptionAmendmentTab status='pending' />,
      },
      {
        label: amendmentTabByStatus.approved,
        content: <ProfileDescriptionAmendmentTab status='approved' />,
      },
      {
        label: amendmentTabByStatus.rejected,
        content: <ProfileDescriptionAmendmentTab status='rejected' />,
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
