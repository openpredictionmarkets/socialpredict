import React from 'react';
import { Link } from 'react-router-dom';

function MarketDetailsTable({ market, numUsers, totalVolume, currentProbability, chartData, formatDateTimeForGrid, formatDateForAxis, isLoggedIn, openBetModal, handleBetAmountChange, betAmount, submitBet, setShowBetModal, openResolveModal, setSelectedResolution, setResolutionPercentage, resolutionPercentage, resolveMarket, isMarketClosed, showBetModal, showResolveModal, isActivityModalOpen, ActivityModalContent, activeTab, changeTab, bets, username }) {
return (
<div>
    <h3 className="text-xs font-medium text-gray-500 uppercase tracking-wider p-6">{market.questionTitle}</h3>
    <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
    <tbody className="bg-primary-background divide-y divide-gray-200">
        <tr>
        <td className="px-6 py-4 text-white">
            <div>
                Link To User
            {/* <Link
                // to={`/user/${market.creatorUsername}`}
                className="text-blue-600 hover:text-blue-800"
            >
                <span role='img' aria-label='Creator'>{creator?.personalEmoji}</span> @{market.creatorUsername}
            </Link> */}
            </div>
        </td>
        <td className="px-6 py-4 text-sm text-gray-500">👤 {/* {numUsers} */}</td>
        <td className="px-6 py-4 text-sm text-gray-500">📊 {/* {totalVolume.toFixed(2)} */}</td>
        <td className="px-6 py-4 text-sm text-gray-500">💬 {/* 0 */}</td>
        <td className="px-6 py-4 text-sm text-gray-500">
            {/* {market.isResolved ? (
            <span>
                RESOLVED: {market.resolutionResult} <p>@ {formatDateTimeForGrid(market.finalResolutionDateTime).toLocaleString()}</p>
            </span>
            ) : (
            <span>Ends: 📅 {formatDateTimeForGrid(market.resolutionDateTime).toLocaleString()}</span>
            )} */}
            Market Resolution Status
        </td>
        </tr>
        <tr>
        <td colSpan='5' className="px-6 py-4 text-sm text-gray-500">
            Probability Area
            {/*
            {market.isResolved ? (
            <>
                <p className="text-left">Final Probability:</p>
                <h2 className="text-left">{currentProbability}</h2>
            </>
            ) : (
            <>
                <p className="text-left">Current Probability:</p>
                <h2 className="text-left">{currentProbability}</h2>
            </>
            )}
            */}
        </td>
        </tr>
        <tr>
        <td colSpan='5'>
            LineChart Here
        </td>
        </tr>
    </tbody>
    </table>
</div>
);
}

export default MarketDetailsTable;