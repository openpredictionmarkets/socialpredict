import React, { useState, useEffect } from 'react';
import { useParams, useHistory } from 'react-router-dom';
import { API_URL } from '../../config';
import { useAuth } from '../../helpers/AuthContent';
import SiteTabs from '../../components/tabs/SiteTabs';
import UserInfoTabContent from '../../components/layouts/profile/public/UserInfoTabContent';
import PortfolioTabContent from '../../components/layouts/profile/public/PortfolioTabContent';
import UserFinancialStatementsLayout from '../../components/layouts/profile/public/UserFinancialStatementsLayout';

const User = () => {
  const [userData, setUserData] = useState(null);
  const { username } = useParams();
  const { isLoggedIn, username: loggedInUsername } = useAuth();
  const history = useHistory();

  useEffect(() => {
    fetch(`${API_URL}/api/v0/userinfo/${username}`)
      .then((response) => response.json())
      .then((data) => setUserData(data))
      .catch((error) => console.error('Error fetching user data:', error));
  }, [username]);

  if (!userData) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-primary-background">
        <div className="text-white text-lg">Loading user profile...</div>
      </div>
    );
  }

  // Define the tabs for the user profile
  const userProfileTabs = [
    {
      label: 'User Info',
      content: <UserInfoTabContent username={username} userData={userData} />
    },
    {
      label: 'Portfolio',
      content: <PortfolioTabContent username={username} />
    },
    {
      label: 'Financials',
      content: <UserFinancialStatementsLayout username={username} />
    }
  ];

  return (
    <div className="min-h-screen bg-primary-background">
      <div className="container mx-auto px-4 py-6">
        <div className="mb-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-white mb-2">
                {userData.personalEmoji} {userData.displayname}
              </h1>
              <div className="text-gray-400">
                @{userData.username}
              </div>
            </div>
            
            {/* Edit Profile Button - Only show if user is viewing their own profile */}
            {isLoggedIn && loggedInUsername === username && (
              <button
                onClick={() => history.push('/profile')}
                className="bg-gold-btn hover:bg-gold-btn-hover text-black font-semibold py-2 px-4 rounded-lg transition-colors duration-200 flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                Edit Profile
              </button>
            )}
          </div>
        </div>
        
        <SiteTabs 
          tabs={userProfileTabs} 
          defaultTab="User Info"
        />
      </div>
    </div>
  );
};

export default User;
