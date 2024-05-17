
import { API_URL } from '../../config';
import React, { useState, useEffect } from 'react';
import PrivateUserLayout from '../../components/layouts/profile/private/PrivateUserLayout'
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout'
import { useAuth } from '../../helpers/AuthContent';

const Profile = () => {

    // Get the logged-in user's ID from context or another state management solution
    const { username, isLoggedIn } = useAuth();

    const [userData, setUserData] = useState(null);


    useEffect(() => {
        fetch(`${API_URL}/api/v0/userinfo/${username}`)
        .then((response) => response.json())
        .then((data) => setUserData(data))
        .catch((error) => console.error('Error fetching user data:', error));
    }, [username]);

    if (!userData) {
        return <div>Loading...</div>;
    }

    return (
        <div>
        <PrivateUserLayout
            userData={userData}
        />
        <PublicUserPortfolioLayout
            username={username}
            userData={userData}
        />
        </div>
    );
};

export default Profile;
