import React from 'react';
import { Link } from 'react-router-dom';

// BetsActivityLayout receives bets from the parent (ActivityTabs) which fetches lazily
// on first tab open. This prevents an eager API call on market details page load.
const BetsActivityLayout = ({ bets, loading, error }) => {
    if (loading) {
        return <div className='p-4 text-gray-400 text-sm'>Loading bets…</div>;
    }

    if (bets === null) {
        return <div className='p-4 text-gray-400 text-sm'>Open this tab to load bets.</div>;
    }

    if (error) {
        return <div className='p-4 text-red-400 text-sm'>{error}</div>;
    }

    if (!bets.length) {
        return <div className='p-4 text-gray-400 text-sm'>No bets placed yet.</div>;
    }

    return (
        <div className='p-4'>
            <div className='sp-grid-bets-header'>
                <div>Username</div>
                <div className='text-center'>Outcome</div>
                <div className='text-right'>Amount</div>
                <div className='text-right'>After</div>
                <div className='text-right'>Placed</div>
            </div>
            {bets.map((bet, index) => (
                <div key={index} className='sp-grid-bets-row mt-2'>
                    <div className='sp-cell-username'>
                        <div className='sp-ellipsis text-xs sm:text-sm font-medium'>
                            <Link
                                to={`/user/${bet.username}`}
                                className='text-blue-500 hover:text-blue-400 transition-colors'
                            >
                                {bet.username}
                            </Link>
                        </div>
                    </div>

                    <div className='justify-self-start sm:justify-self-center'>
                        <span
                            className={`px-2 py-1 rounded text-xs font-bold ${
                                bet.outcome === 'YES' ? 'bg-green-600' : 'bg-red-600'
                            } text-white`}
                        >
                            {bet.outcome}
                        </span>
                    </div>

                    <div className='sp-cell-num text-xs sm:text-sm text-gray-300'>{bet.amount}</div>

                    <div className='hidden sm:block sp-cell-num text-gray-300'>
                        {bet.probability.toFixed(2)}
                    </div>

                    <div className='col-span-3 sm:col-span-1 text-right sp-subline'>
                        {new Date(bet.placedAt).toLocaleString()}
                    </div>
                </div>
            ))}
        </div>
    );
};

export default BetsActivityLayout;
