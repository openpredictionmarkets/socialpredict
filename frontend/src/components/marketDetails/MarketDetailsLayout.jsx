import React from 'react';
import { Link } from 'react-router-dom';
import MarketChart from '../charts/MarketChart';

function MarketDetailsTable(
    { market, numUsers, totalVolume, currentProbability, chartData, formatDateTimeForGrid, formatDateForAxis, isLoggedIn, openBetModal, handleBetAmountChange, betAmount, submitBet, setShowBetModal, openResolveModal, setSelectedResolution, setResolutionPercentage, resolutionPercentage, resolveMarket, isMarketClosed, showBetModal, showResolveModal, isActivityModalOpen, ActivityModalContent, activeTab, changeTab, bets, username }) {
        return (
            <div className="bg-primary-background text-white p-6">
                <h3 className="text-lg font-large tracking-wider">Market Question</h3>
                <div className="divide-y divide-primary-background">
                    <div className="flex">
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">
                        Link To User
                        {/* Uncomment and update the Link component as needed */}
                        {/* <Link
                            to={`/user/${market.creatorUsername}`}
                            className="text-blue-600 hover:text-blue-800"
                        >
                            <span role='img' aria-label='Creator'>{creator?.personalEmoji}</span> @{market.creatorUsername}
                        </Link> */}
                    </div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">👤 {/* {numUsers} */}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">📊 {/* {totalVolume.toFixed(2)} */}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">💬 {/* 0 */}</div>
                    <div className="flex-1 px-6 py-4 text-sm text-gray-500">
                        Market Resolution Status
                        {/* Uncomment and update as needed */}
                        {/* {market.isResolved ? (
                        <span>
                            RESOLVED: {market.resolutionResult} <p>@ {formatDateTimeForGrid(market.finalResolutionDateTime).toLocaleString()}</p>
                        </span>
                        ) : (
                        <span>Ends: 📅 {formatDateTimeForGrid(market.resolutionDateTime).toLocaleString()}</span>
                        )} */}
                    </div>
                    </div>
                    <div className="flex">
                    <div className="flex-1 px-6 py-4 text-lg">
                        Probability Area
                        {/* Uncomment and update as needed */}
                        {/* {market.isResolved ? (
                        <>
                            <p className="text-left">Final Probability:</p>
                            <h2 className="text-left">{currentProbability}</h2>
                        </>
                        ) : (
                        <>
                            <p className="text-left">Current Probability:</p>
                            <h2 className="text-left">{currentProbability}</h2>
                        </>
                        )} */}
                    </div>
                    </div>
                    <div className="flex">
                    <div className="flex-1 px-6 py-4">
                        <MarketChart
                            className="shadow-md border border-custom-gray-light"
                        />
                    </div>
                    </div>
                </div>
                </div>
            );
}

export default MarketDetailsTable;