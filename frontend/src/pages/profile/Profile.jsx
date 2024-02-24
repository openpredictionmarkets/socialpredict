import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import './Profile.css'; // Import the CSS file
import ProfileEdit from '../../components/profileEdit/ProfileEdit';
import { API_URL } from '../../config';

function Profile() {
  const [userData, setUserData] = useState(null);
  const [userPortfolio, setUserPortfolio] = useState([]);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      console.error('Authentication token not found. Please log in again.');
      return;
    }

    // Fetch user data
    fetch(`${API_URL}/api/v0/user/privateprofile`, {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    })
      .then((response) => response.json())
      .then((data) => {
        setUserData(data);
      })
      .catch((error) => console.error('Error fetching user data:', error));
  }, []);

  useEffect(() => {
    if (!userData || !userData.username) {
      return;
    }

    const token = localStorage.getItem('token');
    if (!token) {
      console.error('Authentication token not found.');
      return;
    }

    // Fetch user portfolio
    fetch(`${API_URL}/api/v0/portfolio/${userData.username}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((response) => response.json())
      .then((data) => setUserPortfolio(data))
      .catch((error) => console.error('Error fetching user portfolio:', error));
  }, [userData]); // This useEffect gets triggered when userData is set

  // Function to render the user's portfolio
  const renderUserPortfolio = () => {
    return userPortfolio.map((portfolioItem, index) => (
      <tr key={index} className='portfolio-center-content-table'>
        <td>
          <Link to={`/markets/${portfolioItem.marketId}`}>
            {portfolioItem.questionTitle}
          </Link>
        </td>
        <td>Total YES Bets: ${portfolioItem.totalYesBets}</td>
        <td>Total NO Bets: ${portfolioItem.totalNoBets}</td>
      </tr>
    ));
  };

  // Function to render personal links as clickable links
  const renderPersonalLinks = () => {
    const linkKeys = [
      'personalLink1',
      'personalLink2',
      'personalLink3',
      'personalLink4',
    ];
    return linkKeys.map((key) =>
      userData[key] ? (
        <div key={key} className='nav-link-callout'>
          <a
            className='nav-link'
            href={userData[key]}
            target='_blank'
            rel='noopener noreferrer'
          >
            {userData[key]}
          </a>
        </div>
      ) : null
    );
  };

  if (!userData) {
    return <div>Loading...</div>;
  }

  return (
    <div className='profile'>
      <table>
        <tbody>
          <tr>
            <td colSpan='2' className='header'>
              <h3>Profile Settings</h3>
            </td>
          </tr>
          <tr>
            <td className='label'>Username (Permanent):</td>
            <td>{userData.username}</td>
          </tr>
          <tr>
            <td className='label'>Personal Emoji:</td>
            <td>
              {userData.personalEmoji && (
                <div className='profile-cell-container'>
                  <span className='profile-emoji'>
                    {userData.personalEmoji}
                  </span>
                  <button className='edit-button'>Edit</button>
                </div>
              )}
            </td>
          </tr>
          <tr>
            <td className='label'>Display Name:</td>
            <td>
              {userData.personalEmoji && (
                <div className='profile-cell-container'>
                  <span>{userData.displayname}</span>
                  <button className='edit-button'>Edit</button>
                </div>
              )}
            </td>
          </tr>
          <tr>
            <td className='label'>Description:</td>
            <td>{userData.description}</td>
          </tr>
          <tr>
            <td className='label'>Personal Links:</td>
            <td>{renderPersonalLinks()}</td>
          </tr>
          <tr>
            <td colSpan='2' className='header'>
              <h3>Private Info</h3>
            </td>
          </tr>
          <tr>
            <td>Email (Fake for Now, No Update Capability):</td>
            <td>{userData.email}</td>
          </tr>
          <tr>
            <td>API Key (Not Valid Yet):</td>
            <td>{userData.apiKey}</td>
          </tr>
          <tr>
            <td colSpan='2' className='header'>
              <h3>Financial Stats</h3>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Account Balance: ${userData.accountBalance.toFixed(2)}
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Maximum Debt Allowed:
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              <b>Balance Sheet</b>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Retained Earnings: $E
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Amount In Play: $X
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Amount Borrowed: $Y
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Equity = Retained Earnings + Amount In Play - Amount Borrowed
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              <b>Income Statement</b>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Trading Profits: ${userData.accountBalance.toFixed(2)}
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Work Profits: $W
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              Total Profits: Trading Profits + $W
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              <b>Statment of Retained Earnings</b>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='account-balance'>
              <b>Cash Flow Statement</b>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='header'>
              <h3>Portfolio</h3>
            </td>
          </tr>
          <tr>
            <td colSpan='2' className='header'>
              Markets Invested In: {renderUserPortfolio()}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export default Profile;
