import { API_URL } from '../../../../config';
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom'; // Import Link
import { unwrapApiResponse } from '../../../../utils/apiResponse';

const paginationButtonClass = [
    'rounded',
    'border',
    'border-transparent',
    'bg-neutral-btn',
    'px-3',
    'py-1.5',
    'text-xs',
    'font-semibold',
    'text-white',
    'transition-colors',
    'duration-200',
    'hover:bg-neutral-btn-hover',
    'disabled:cursor-not-allowed',
    'disabled:bg-custom-gray-light',
    'disabled:text-gray-400',
    'disabled:opacity-60',
].join(' ');

const BetsActivityLayout = ({ marketId, refreshTrigger }) => {
    const pageSize = 20;
    const [bets, setBets] = useState([]);
    const [page, setPage] = useState(0);

    useEffect(() => {
        const fetchBets = async () => {
            const response = await fetch(`${API_URL}/v0/markets/bets/${marketId}`, {
            });
            if (response.ok) {
                const data = unwrapApiResponse(await response.json());
                setBets(data
                    .sort((a, b) => new Date(b.placedAt) - new Date(a.placedAt)));
                setPage(0);
            } else {
                console.error('Error fetching bets:', response.statusText);
            }
        };
        fetchBets();
    }, [marketId, refreshTrigger]);

    const pageStart = page * pageSize;
    const visibleBets = bets.slice(pageStart, pageStart + pageSize);
    const canPageBack = page > 0;
    const canPageForward = pageStart + pageSize < bets.length;

    return (
        <div className="p-4">
            <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div className="text-xs uppercase tracking-[0.16em] text-gray-400">
                    Showing bets {bets.length ? pageStart + 1 : 0}-{Math.min(pageStart + pageSize, bets.length)} of {bets.length}
                </div>
                <div className="flex gap-2">
                    <button
                        type="button"
                        onClick={() => setPage(current => Math.max(0, current - 1))}
                        disabled={!canPageBack}
                        className={paginationButtonClass}
                    >
                        Previous
                    </button>
                    <button
                        type="button"
                        onClick={() => setPage(current => current + 1)}
                        disabled={!canPageForward}
                        className={paginationButtonClass}
                    >
                        Next
                    </button>
                </div>
            </div>
            <div className="sp-grid-bets-header">
                <div>Username</div>
                <div className="text-center">Outcome</div>
                <div className="text-right">Amount</div>
                <div className="text-right">After</div>
                <div className="text-right">Placed</div>
            </div>
            {visibleBets.map((bet, index) => (
                <div key={index} className="sp-grid-bets-row mt-2">
                    {/* Username */}
                    <div className="sp-cell-username">
                        <div className="sp-ellipsis text-xs sm:text-sm font-medium">
                            <Link to={`/user/${bet.username}`} className="text-blue-500 hover:text-blue-400 transition-colors">
                                {bet.username}
                            </Link>
                        </div>
                    </div>

                    {/* Outcome */}
                    <div className="justify-self-start sm:justify-self-center">
                        <span className={`px-2 py-1 rounded text-xs font-bold ${bet.outcome === 'YES' ? 'bg-green-600' : 'bg-red-600'} text-white`}>
                            {bet.outcome}
                        </span>
                    </div>

                    {/* Amount */}
                    <div className="sp-cell-num text-xs sm:text-sm text-gray-300">{bet.amount}</div>

                    {/* After (sm+) */}
                    <div className="hidden sm:block sp-cell-num text-gray-300">{bet.probability.toFixed(2)}</div>

                    {/* Placed (stack full width on xs) */}
                    <div className="col-span-3 sm:col-span-1 text-right sp-subline">
                        {new Date(bet.placedAt).toLocaleString()}
                    </div>
                </div>
            ))}
        </div>
    );


};

export default BetsActivityLayout;
