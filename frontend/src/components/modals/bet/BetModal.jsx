import React, { useState } from 'react';
import { BetButton } from '../../buttons/trade/BetButtons';
import TradeTabs from '../../tabs/TradeTabs';
import { submitBet } from './BetUtils'

const BetModalButton = ({ marketId, token, onTransactionSuccess, currentProbability, totalYes, totalNo }) => {
    const [showBetModal, setShowBetModal] = useState(false);
    const toggleBetModal = () => setShowBetModal(prev => !prev);

    const handleTransactionSuccess = () => {
        setShowBetModal(false);  // Close modal
        if (onTransactionSuccess) {
            onTransactionSuccess();  // Trigger data refresh
        }
    };

    return (
        <div>
            <BetButton onClick={toggleBetModal} className="ml-6 w-10%" />
            {showBetModal && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
                    <div className="bet-modal relative bg-blue-900 p-6 rounded-lg text-white m-6 mx-auto" style={{ width: '350px' }}>

                    <TradeTabs
                        marketId={marketId}
                        token={token}
                        onTransactionSuccess={handleTransactionSuccess}
                        currentProbability={currentProbability}
                        totalYes={totalYes}
                        totalNo={totalNo}                        
                    />

                        <button onClick={toggleBetModal} className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white">
                            ✕
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default BetModalButton;
