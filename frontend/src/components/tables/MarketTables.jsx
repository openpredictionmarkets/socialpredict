import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';

const TableHeader = () => (
  <thead className='bg-gray-900'>
    <tr>
      {[
        'Trade',
        '🪙',
        'Question',
        '📅 Closes',
        'Creator',
        '👤 Users',
        '📊 Size',
        '💬',
      ].map((header, index) => (
        <th
          key={index}
          className='px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider'
        >
          {header}
        </th>
      ))}
    </tr>
  </thead>
);

const formatResolutionDate = (resolutionDateTime) => {
  const now = new Date();
  const resolutionDate = new Date(resolutionDateTime);

  return resolutionDate < now ? 'Closed' : resolutionDate.toLocaleDateString();
};

const MarketRow = ({ marketData }) => (
  <tr className='hover:bg-gray-700 transition-colors duration-200'>
    <td className='px-6 py-4 whitespace-nowrap'>
      <Link
        to={`/markets/${marketData.market.id}`}
        className='text-blue-400 hover:text-blue-300'
      >
        ⬆️⬇️
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-300'>
      {marketData.lastProbability.toFixed(3)}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-300'>
      <Link
        to={`/markets/${marketData.market.id}`}
        className='hover:text-blue-400 transition-colors duration-200'
      >
        {marketData.market.questionTitle}
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {formatResolutionDate(marketData.market.resolutionDateTime)}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      <Link
        to={`/user/${marketData.creator.username}`}
        className='flex items-center hover:text-blue-400 transition-colors duration-200'
      >
        <span role='img' aria-label='Creator' className='mr-2'>
          {marketData.creator.personalEmoji}
        </span>
        @{marketData.creator.username}
      </Link>
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.numUsers}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.totalVolume}
    </td>
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>0</td>
  </tr>
);

function MarketsTable() {
  const [marketsData, setMarketsData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchMarkets = async () => {
      try {
        const response = await fetch(`${API_URL}/api/v0/markets`);
        if (!response.ok) throw new Error('Failed to fetch markets');
        const data = await response.json();
        setMarketsData(data.markets);
      } catch (error) {
        console.error('Error fetching market data:', error);
        setError(error.toString());
      } finally {
        setLoading(false);
      }
    };

    fetchMarkets();
  }, []);

  if (loading) return <div className='p-4 text-center'>Loading markets...</div>;
  if (error)
    return <div className='p-4 text-center text-red-500'>Error: {error}</div>;
  if (!marketsData || marketsData.length === 0)
    return <div className='p-4 text-center'>No markets found.</div>;

  return (
    <div className='w-full max-w-7xl mx-auto p-1 sm:px-6 lg:px-8'>
      <h1 className='text-2xl font-semibold text-gray-300 mb-6'>Markets</h1>
      <div className='bg-gray-800 shadow-md rounded-lg overflow-hidden'>
        <div className='overflow-x-auto'>
          <table className='min-w-full divide-y divide-gray-700'>
            <TableHeader />
            <tbody className='bg-gray-800 divide-y divide-gray-700'>
              {marketsData.map((marketData, index) => (
                <MarketRow key={index} marketData={marketData} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default MarketsTable;
