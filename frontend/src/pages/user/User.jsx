import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { API_URL } from '../../config';
import PublicUserInfoLayout from '../../components/layouts/profile/public/PublicUserInfoLayout'
import PublicUserPortfolioLayout from '../../components/layouts/profile/public/PublicUserPortfolioLayout'

const User = () => {
  const [userData, setUserData] = useState(null);
  const { username } = useParams();

  useEffect(() => {
    fetch(`${API_URL}/v0/userinfo/${username}`)
      .then((response) => response.json())
      .then((data) => setUserData(data))
      .catch((error) => console.error('Error fetching user data:', error));
  }, [username]);

  if (!userData) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <PublicUserInfoLayout userData={userData} />
      <PublicUserPortfolioLayout username={username} userData={userData} />
    </div>
  );
};

export default User;
