import React, { useState } from 'react';
import { BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/BetButtons';
import { submitBet } from './TradeUtils'


const BuySharesLayout = ({ marketId, token, onTransactionSuccess }) => {
    const [betAmount, setBetAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);

    const handleBetAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10);
        setBetAmount(newAmount >= 0 ? newAmount : '');
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

        submitBet(betData, token, (data) => {
            alert(`Bet placed successfully! Bet ID: ${data.id}`);
            onTransactionSuccess();
        }, (error) => {
            alert(`Error placing bet: ${error.message}`);
        });
    };

    return (
        <div className="p-6 bg-blue-900 rounded-lg text-white">
            <h2 className="text-xl mb-4">Purchase Shares</h2>
            <div className="flex justify-center space-x-4 mb-4">
                <BetNoButton onClick={() => setSelectedOutcome('NO')} />
                <BetYesButton onClick={() => setSelectedOutcome('YES')} />
            </div>
            <div className="border-t border-gray-200 my-2"></div>
            <div className="flex items-center space-x-4 mb-4">
                <h2 className="text-xl">Amount</h2>
                <BetInputAmount value={betAmount} onChange={handleBetAmountChange} />
            </div>
            <ConfirmBetButton onClick={handleBetSubmission} selectedDirection={selectedOutcome} />
        </div>
    );
};

export default BuySharesLayout;