import { useAuth } from '../../helpers/AuthContent';
import PrivateUserInfoLayout from '../../components/layouts/profile/private/PrivateUserInfoLayout';
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout';
import useUserData from '../../hooks/useUserData';
import usePortfolio from '../../hooks/usePortfolio';

const LoadingSpinner = () => (
  <div className='flex justify-center items-center h-64'>
    <div className='animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-blue-500'></div>
  </div>
);

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
  <div className='bg-primary-background shadow-md rounded-lg p-6 text-center text-gray-600'>
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
    <div className='flex flex-col space-y-6 min-h-screen p-6 bg-gray-100'>
      <PrivateUserInfoLayout userData={userData} />
      {portfolio && portfolio.portfolioItems.length > 0 ? (
        <PublicUserPortfolioLayout username={username} userData={userData} />
      ) : (
        <NoPortfolio />
      )}
    </div>
  );
};

export default Profile;
