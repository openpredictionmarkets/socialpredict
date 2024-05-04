import React, { useState, useEffect } from 'react';
import { BetButton, BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/BetButtons';
import { submitBet } from './BetUtils'

const BetModalButton = ({ marketId, token, onBetSuccess }) => {
    const [showBetModal, setShowBetModal] = useState(false);
    const [betAmount, setBetAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);

    const toggleBetModal = () => setShowBetModal(prev => !prev);

    const handleBetAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10);
        if (!isNaN(newAmount) && newAmount >= 0) {
            setBetAmount(newAmount);
        } else {
            setBetAmount('');
        }
    };

    const handleBetSubmission = () => {
        if (!token) {
            alert('Please log in to place a bet.');
            return;
        }

        const betData = {
            marketId,
            amount: betAmount,
            outcome: selectedOutcome,
        };

        submitBet(betData, token,
            (data) => {
                alert(`Bet placed successfully! Bet ID: ${data.id}`);
                setShowBetModal(false);
                onBetSuccess();
            },
            (error) => {
                alert(`Error placing bet: ${error.message}`);
            }
        );
    };

    return (
        <div>
            <BetButton onClick={toggleBetModal} className="ml-6 w-10%" />
            {showBetModal && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
                    <div className="bet-modal relative bg-blue-900 p-6 rounded-lg text-white m-6 mx-auto" style={{ width: '350px' }}>
                        <h2 className="text-xl mb-4">Purchase Shares</h2>

                        <div className="flex justify-center space-x-4 mb-4">
                            <div>
                                <BetNoButton onClick={() => setSelectedOutcome('NO')} />
                            </div>
                            <div>
                                <BetYesButton onClick={() => setSelectedOutcome('YES')} />
                            </div>
                        </div>

                        <div className="border-t border-gray-200 my-2"></div>

                        <div className="flex items-center space-x-4 mb-4">
                            <h2 className="text-xl">Amount</h2>
                            <BetInputAmount value={betAmount} onChange={handleBetAmountChange} />
                        </div>

                        <div className="mt-4">
                            <ConfirmBetButton onClick={handleBetSubmission} selectedDirection={selectedOutcome} />
                        </div>

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
