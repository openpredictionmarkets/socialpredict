import React from 'react';
import { Link } from 'react-router-dom';
import MarketChart from '../charts/MarketChart';
import { formatDateTimeForGrid } from '../utils/dateTimeTools/FormDateTimeTools'

function MarketDetailsTable({ market, creator, probabilityChanges, numUsers, totalVolume }) {

    return (
        <div className="bg-primary-background text-white p-6">
            <h3 className="text-lg font-large tracking-wider">{market.questionTitle}</h3>
            <div className="divide-y divide-primary-background">
                <div className="flex">
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">
                        <Link
                            to={`/user/${market.creatorUsername}`}
                            className="text-blue-600 hover:text-blue-800"
                        >
                            <span role='img' aria-label='Creator'>{creator.personalEmoji}</span> @{market.creatorUsername}
                        </Link>
                    </div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">👤 {numUsers}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">📊 {totalVolume.toFixed(2)}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">💬 {/* 0 */}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">
                        {market.isResolved ? (
                            <span>
                                RESOLVED: {market.resolutionResult} <p>@ {formatDateTimeForGrid(market.finalResolutionDateTime).toLocaleString()}</p>
                            </span>
                        ) : (
                            <>
                                <div>CLOSE DATE:</div>
                                <div>📅 {formatDateTimeForGrid(market.resolutionDateTime).toLocaleString()}</div>
                            </>
                        )}
                    </div>
                </div>
                <div className="flex">
                    <div className="flex-1 px-6 py-4">
                        <MarketChart
                            className="shadow-md border border-custom-gray-light"
                            data={probabilityChanges}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
}

export default MarketDetailsTable;
