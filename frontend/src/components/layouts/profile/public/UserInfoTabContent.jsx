import React, { useState, useEffect } from 'react';
import { API_URL } from '../../../../config';
import LoadingSpinner from '../../../loaders/LoadingSpinner';

const UserInfoTabContent = ({ username, userData }) => {
    const [userCredit, setUserCredit] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchUserCredit = async () => {
            try {
                console.log(`Fetching user credit for: ${username} from ${API_URL}/v0/usercredit/${username}`);
                const response = await fetch(`${API_URL}/v0/usercredit/${username}`);
                if (response.ok) {
                    const data = await response.json();
                    console.log('User credit data:', data);
                    setUserCredit(data);
                } else {
                    throw new Error(`Error fetching user credit: ${response.statusText}`);
                }
            } catch (err) {
                console.error('Error fetching user credit:', err);
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        if (username) {
            fetchUserCredit();
        }
    }, [username]);

    const renderPersonalLinks = () => {
        if (!userData) return null;

        const linkKeys = ['personalink1', 'personalink2', 'personalink3', 'personalink4'];
        return linkKeys.map(key => {
            const link = userData[key];
            return link ? (
                <div key={key} className='nav-link text-info-blue hover:text-blue-800'>
                    <a
                        href={link}
                        target='_blank'
                        rel='noopener noreferrer'
                    >
                        {link}
                    </a>
                </div>
            ) : null;
        });
    };

    if (!userData) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="flex items-center justify-center">
                    <LoadingSpinner />
                    <span className="ml-2 text-gray-300">Loading user info...</span>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Credits Section - Prominent Display */}
            <div className="bg-primary-background shadow-md rounded-lg border-2 border-gold-btn">
                <div className="p-6">
                    <h3 className="text-xl font-bold text-gold-btn mb-4 text-center">Credits Available</h3>
                    <div className="flex items-center justify-center">
                        {loading ? (
                            <div className="flex items-center">
                                <LoadingSpinner />
                                <span className="ml-2 text-gray-300">Loading...</span>
                            </div>
                        ) : error ? (
                            <div className="text-red-400">
                                Error loading credit: {error}
                            </div>
                        ) : (
                            <div className="text-center">
                                <div className="text-4xl font-bold text-green-400">
                                    {userCredit?.credit ?? 'N/A'} 🪙
                                </div>
                                <div className="text-sm text-gray-400 mt-2">
                                    Available spending money
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* User Info Section */}
            <div className="bg-primary-background shadow-md rounded-lg">
                <div className="p-6">
                    <h3 className="text-lg font-medium text-custom-gray-verylight mb-4">User Info</h3>
                    <table className="min-w-full divide-y divide-custom-gray-dark">
                        <tbody className="bg-primary-background divide-y divide-custom-gray-dark">
                            <tr>
                                <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Username (Permanent):</td>
                                <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.username}</td>
                            </tr>
                            <tr>
                                <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Personal Emoji:</td>
                                <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.personalEmoji}</td>
                            </tr>
                            <tr>
                                <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Display Name:</td>
                                <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.displayname}</td>
                            </tr>
                            <tr>
                                <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Description:</td>
                                <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.description}</td>
                            </tr>
                            <tr>
                                <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Personal Links:</td>
                                <td className="px-6 py-4 text-sm text-custom-gray-light">{renderPersonalLinks()}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default UserInfoTabContent;
