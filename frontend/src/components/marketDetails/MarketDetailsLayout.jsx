import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import MarketChart from '../charts/MarketChart';
import { formatDateTimeForGrid } from '../utils/dateTimeTools/FormDateTimeTools';
import DescriptionButton from '../buttons/DescriptionButton';

function MarketDetailsTable({ market, creator, numUsers, totalVolume, currentProbability, probabilityChanges }) {
    const [showFullDescription, setShowFullDescription] = useState(false);

    // Function to truncate the description to 50 characters
    const truncateDescription = (description) => {
        if (description.length > 50) {
            return `${description.substring(0, 100)}...`;
        }
        return description;
    };

    return (
        <div className="bg-primary-background text-white p-6">
            <h3 className="text-lg font-large tracking-wider">{market.questionTitle}</h3>
            <div className="divide-y divide-primary-background">
                <div className="flex justify-between items-start">
                    {/* Left section for user, icon and current probability */}
                    <div className="space-y-4">
                        <div className="text-sm text-gray-500">
                            <Link
                                to={`/user/${market.creatorUsername}`}
                                className="text-blue-600 hover:text-blue-800 flex items-center"
                            >
                                <span role='img' aria-label='Creator'>{creator.personalEmoji}</span>
                                @{market.creatorUsername}
                            </Link>
                        </div>
                        <div className="text-2xl font-semibold">🪙 {currentProbability}</div>
                    </div>
                    {/* Right section for details, inline and tightly grouped */}
                    <div className="flex space-x-4 items-center">
                        <div className="text-sm text-gray-500">👤 {numUsers}</div>
                        <div className="text-sm text-gray-500">📊 {totalVolume.toFixed(2)}</div>
                        <div className="text-sm text-gray-500">💬 0 </div>
                        <div className="text-sm text-gray-500">
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
                </div>
                </div>
                <div className="px-6 py-4">
                    <MarketChart
                        data={probabilityChanges}
                        currentProbability={currentProbability}
                        className="shadow-md border border-custom-gray-light"
                    />
                </div>
                <div className="px-6 py-4">
                    <DescriptionButton onClick={() => setShowFullDescription(!showFullDescription)}>
                            {showFullDescription ? 'Hide Description' : 'Show Full Description'}
                    </DescriptionButton>
                    <div className="mt-4 bg-gray-700 p-4 rounded text-gray-200">
                            <p>{showFullDescription ? market.description : truncateDescription(market.description)}</p>
                    </div>
                </div>
            </div>
    );
}

export default MarketDetailsTable;
