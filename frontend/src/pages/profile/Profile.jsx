import React from 'react';
import { useAuth } from '../../helpers/AuthContent';
import PrivateUserInfoLayout from '../../components/layouts/profile/private/PrivateUserInfoLayout';
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout';
import useUserData from '../../hooks/useUserData';
import usePortfolio from '../../hooks/usePortfolio';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';

const ErrorMessage = ({ message }) => (
  <div
    className='bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative'
    role='alert'
  >
    <strong className='font-bold'>Error:</strong>
    <span className='block sm:inline'> {message}</span>
  </div>
);

const NoPortfolio = () => (
  <div className='bg-gray-800 shadow-md rounded-lg p-6 text-center text-gray-300'>
    No portfolio found. User has likely not made any trades yet.
  </div>
);

const Profile = () => {
  const { username } = useAuth();
  const { userData, userLoading, userError } = useUserData(username);
  const { portfolio, portfolioLoading, portfolioError } =
    usePortfolio(username);

  if (userLoading || portfolioLoading) {
    return <LoadingSpinner />;
  }

  if (userError) {
    return <ErrorMessage message={`Error loading user data: ${userError}`} />;
  }

  if (portfolioError) {
    return (
      <ErrorMessage message={`Error loading portfolio: ${portfolioError}`} />
    );
  }

  return (
    <div className='flex flex-col space-y-6  text-white'>
      <h1 className='text-2xl font-bold mb-4'>Profile Details & Edit</h1>
      <div className='bg-gray-800 rounded-lg p-6 shadow-lg'>
        <PrivateUserInfoLayout userData={userData} />
      </div>
      <div className='bg-gray-800 rounded-lg p-6 shadow-lg'>
        {portfolio && portfolio.portfolioItems.length > 0 ? (
          <PublicUserPortfolioLayout username={username} userData={userData} />
        ) : (
          <NoPortfolio />
        )}
      </div>
    </div>
  );
};

export default Profile;
