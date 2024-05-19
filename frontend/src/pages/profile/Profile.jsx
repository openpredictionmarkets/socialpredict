import { API_URL } from '../../config';
import React, { useState, useEffect } from 'react';
import PrivateUserLayout from '../../components/layouts/profile/private/PrivateUserLayout';
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout';
import { useAuth } from '../../helpers/AuthContent';

const Profile = () => {
    const { username, isLoggedIn } = useAuth();
    const [userData, setUserData] = useState(null);
    const [portfolioTotal, setPortfolioTotal] = useState({
        portfolioItems: [],
        totalSharesOwned: 0,
    });

    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/userinfo/${username}`);
                const data = await response.json();
                setUserData(data);
            } catch (error) {
                console.error('Error fetching user data:', error);
            }
        };

        if (username) {
            fetchUserData();
        }
    }, [username]);

    useEffect(() => {
        const fetchPortfolio = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/portfolio/${username}`);
                const data = await response.json();
                setPortfolioTotal(data);
            } catch (error) {
                console.error('Error fetching portfolio:', error);
            }
        };

        if (username) {
            fetchPortfolio();
        }
    }, [username]);

    if (!userData) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            <PrivateUserLayout userData={userData} />
            {portfolioTotal.portfolioItems.length > 0 && (
                <PublicUserPortfolioLayout username={username} userData={userData} />
            )}
        </div>
    );
};

export default Profile;
