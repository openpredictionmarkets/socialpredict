import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';
import PrivateUserInfoLayout from '../../components/layouts/profile/private/PrivateUserInfoLayout';
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout';
import { useAuth } from '../../helpers/AuthContent';

const Profile = () => {
    const { username } = useAuth();
    const [userData, setUserData] = useState({ data: null, loading: true, error: null });
    const [portfolio, setPortfolio] = useState({ data: { portfolioItems: [] }, loading: true, error: null });

    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/userinfo/${username}`);
                if (response.ok) {
                    const data = await response.json();
                    setUserData({ data, loading: false, error: null });
                } else {
                    throw new Error('Failed to fetch user data');
                }
            } catch (error) {
                console.error('Error fetching user data:', error);
                setUserData({ data: null, loading: false, error: error.toString() });
            }
        };

        const fetchPortfolio = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/portfolio/${username}`);
                if (response.ok) {
                    const data = await response.json();
                    setPortfolio({ data: { ...data, portfolioItems: data.portfolioItems || [] }, loading: false, error: null });
                } else {
                    throw new Error('Failed to fetch portfolio');
                }
            } catch (error) {
                console.error('Error fetching portfolio:', error);
                setPortfolio({ data: { portfolioItems: [] }, loading: false, error: error.toString() });
            }
        };

        if (username) {
            fetchUserData();
            fetchPortfolio();
        }
    }, [username]);

    if (userData.loading || portfolio.loading) {
        return <div>Loading...</div>;
    }

    if (userData.error) {
        return <div>Error loading user data: {userData.error}</div>;
    }

    return (
        <div className="flex-col min-h-screen">
            <PrivateUserInfoLayout userData={userData.data} />
            {portfolio.data && portfolio.data.portfolioItems.length > 0 ? (
                <PublicUserPortfolioLayout username={username} userData={userData.data} />
            ) : (
                <div className="bg-primary-background shadow-md rounded-lg p-6 text-center">
                    No portfolio found. User has likely not made any trades yet.
                </div>
            )}
        </div>
    );
};

export default Profile;
