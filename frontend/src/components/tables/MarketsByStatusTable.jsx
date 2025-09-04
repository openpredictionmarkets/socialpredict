import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { API_URL } from '../../config';
import formatResolutionDate from '../../helpers/formatResolutionDate';
import MobileMarketCard from './MobileMarketCard';
import LoadingSpinner from '../loaders/LoadingSpinner';
import ExpandableLink from '../utils/ExpandableLink';

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
        'Resolution',
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
    <td className='px-6 py-4 text-sm font-medium text-gray-300'>
      <ExpandableLink
        text={marketData.market.questionTitle}
        to={`/markets/${marketData.market.id}`}
        maxLength={45}
        className=""
        linkClassName="hover:text-blue-400 transition-colors duration-200"
        buttonClassName="text-xs text-blue-400 hover:text-blue-300 transition-colors ml-1"
        expandIcon="📐"
      />
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
    <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-400'>
      {marketData.market.isResolved ? (
        <span
          className={
            marketData.market.resolutionResult === 'YES'
              ? 'text-green-400'
              : 'text-red-400'
          }
        >
          {marketData.market.resolutionResult === 'YES'
            ? 'Resolved YES'
            : 'Resolved NO'}
        </span>
      ) : (
        'Pending'
      )}
    </td>
  </tr>
);

function MarketsByStatusTable({ status }) {
  const [marketsData, setMarketsData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchMarkets = async () => {
      setLoading(true);
      setError('');
      
      try {
        const endpoint = status === 'all' 
          ? `${API_URL}/api/v0/markets`
          : `${API_URL}/api/v0/markets/${status}`;
        
        const response = await fetch(endpoint);
        if (!response.ok) throw new Error(`Failed to fetch ${status} markets`);
        
        const data = await response.json();
        
        // Handle different response structures
        if (status === 'all') {
          setMarketsData(data.markets || []);
        } else {
          setMarketsData(data.markets || []);
        }
      } catch (error) {
        console.error(`Error fetching ${status} market data:`, error);
        setError(error.toString());
      } finally {
        setTimeout(() => setLoading(false), 300);
      }
    };

    fetchMarkets();
  }, [status]);

  if (loading)
    return (
      <div className='p-4 text-center'>
        <LoadingSpinner />
        Loading {status} markets...
      </div>
    );
    
  if (error)
    return <div className='p-4 text-center text-red-500'>Error: {error}</div>;

  return (
    <div className='w-full'>
      {marketsData.length === 0 ? (
        <div className='p-4 text-center text-gray-400'>
          No {status} markets found.
        </div>
      ) : (
        <>
          <div className='md:hidden'>
            {marketsData.map((marketData, index) => (
              <MobileMarketCard key={index} marketData={marketData} />
            ))}
          </div>
          <div className='hidden md:block bg-gray-800 shadow-md rounded-lg overflow-hidden'>
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
        </>
      )}
    </div>
  );
}

export default MarketsByStatusTable;
