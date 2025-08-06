import React, { useState, useEffect } from 'react';
import { API_URL } from '../../../../config';
import { Link } from 'react-router-dom';

const PublicUserPortfolioLayout = ({ username, userData }) => {
    const [portfolioTotal, setPortfolioTotal] = useState({
        portfolioItems: [],
        totalSharesOwned: 0,
    });

    useEffect(() => {
        const fetchPortfolio = async () => {
            console.log(`Fetching portfolio for user: ${username} from ${API_URL}/api/v0/portfolio/${username}`);
            const response = await fetch(`${API_URL}/api/v0/portfolio/${username}`);
            if (response.ok) {
                const data = await response.json();
                console.log('Portfolio data:', data);
                setPortfolioTotal(data);
            } else {
                console.error('Error fetching portfolio:', response.statusText);
            }
        };

        if (username) {
            fetchPortfolio();
        }
    }, [username]);

    const { portfolioItems, totalSharesOwned } = portfolioTotal;

    if (!portfolioItems || !portfolioItems.length) {
        return <div className="bg-primary-background shadow-md rounded-lg">
            <div className="p-6">
                <table className="w-full">
                    <tbody>
                        <tr>
                            <td>
                                No portfolio found.
                            </td>
                        </tr>
                        <tr>
                            <td>
                                User has likely not made any trades yet.
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>;
    }

    return (
        <div className="bg-primary-background shadow-md rounded-lg">
            <div className="p-6">
                <h3 className="text-lg font-medium text-gray-700 mb-2">Account Balance</h3>
                <p><strong>Account Balance:</strong> {userData.accountBalance}</p>
                <p><strong>Initial Account Balance:</strong> {userData.initialAccountBalance}</p>
            </div>
            <div className="p-6">
                <h3 className="text-lg font-medium text-gray-700 mb-2">Portfolio Value</h3>
                <p><strong>Total Shares Owned:</strong> {totalSharesOwned}</p>
            </div>
            <div className="p-6">
                <h3 className="text-lg font-medium text-gray-700 mb-2">Portfolio</h3>
                <table className="w-full divide-y divide-gray-200 bg-primary-background">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Question</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total YES Shares</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total NO Shares</th>
                        </tr>
                    </thead>
                    <tbody>
                        {portfolioItems.map((portfolioItem, index) => (
                            <tr key={index} className='bg-primary-background divide-y divide-gray-200'>
                                <td className="px-6 py-4 text-sm font-mono text-gray-500">
                                    <Link to={`/markets/${portfolioItem.marketId}`} className="text-blue-600 hover:text-blue-800">
                                        {portfolioItem.questionTitle}
                                    </Link>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-500">{portfolioItem.yesSharesOwned}</td>
                                <td className="px-6 py-4 text-sm text-gray-500">{portfolioItem.noSharesOwned}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default PublicUserPortfolioLayout;
