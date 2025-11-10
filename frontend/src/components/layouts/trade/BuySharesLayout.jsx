import React, { useState, useEffect } from 'react';
import { BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/trade/BetButtons';
import MarketProjectionLayout from '../marketprojection/MarketProjectionLayout';
import { submitBet } from './TradeUtils';
import { useMarketLabels } from '../../../hooks/useMarketLabels';
import { API_URL } from '../../../config';


const BuySharesLayout = ({ marketId, market, token, onTransactionSuccess }) => {
    const [betAmount, setBetAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);
    const [feeData, setFeeData] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    
    // Get custom labels for this market
    const { yesLabel, noLabel } = useMarketLabels(market);

    useEffect(() => {
        const fetchFeeData = async () => {
            if (!token) {
                setIsLoading(false);
                return;
            }
            try {
                const response = await fetch(`${API_URL}/v0/setup`, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });
                if (!response.ok) {
                    throw new Error(`Failed to load setup: ${response.status}`);
                }
                const data = await response.json();
                setFeeData(data.Betting.BetFees);
            } catch (error) {
                console.error('Error fetching fee data:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchFeeData();
    }, [token]);


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
            alert(`Bet placed successfully! ${data.amount} on ${data.outcome}`);
            onTransactionSuccess();
        }, (error) => {
            alert(`Error placing bet: ${error.message}`);
        });
    };

    return (
        <div className="p-6 bg-blue-900 rounded-lg text-white">
            <h2 className="text-xl mb-4">Purchase Shares</h2>
            <div className="flex justify-center space-x-4 mb-4">
                <BetNoButton 
                    onClick={() => setSelectedOutcome('NO')} 
                    label={noLabel}
                />
                <BetYesButton 
                    onClick={() => setSelectedOutcome('YES')} 
                    label={yesLabel}
                />
            </div>
            <div className="border-t border-gray-200 my-2"></div>
            <div className="flex items-center space-x-4 mb-4">
                <h2 className="text-xl">Amount</h2>
                <BetInputAmount value={betAmount} onChange={handleBetAmountChange} />
            </div>
            <ConfirmBetButton onClick={handleBetSubmission} selectedDirection={selectedOutcome} yesLabel={yesLabel} noLabel={noLabel} />
            <div>
            <div className="border-t border-gray-200 my-2"></div>

            {!isLoading && feeData && (
                <div className="mb-4">
                    {feeData.InitialBetFee === 0 && feeData.BuySharesFee === 0 ? (
                        <p className="text-sm text-gray-300">No fees</p>
                    ) : (
                        <>
                            {feeData.InitialBetFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Initial Trade Fee: {feeData.InitialBetFee}
                                    <span className="block">Does not apply if already traded on this market.</span>
                                </p>
                            )}
                            {feeData.BuySharesFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Trading Fee (Buying Share): {feeData.BuySharesFee}
                                </p>
                            )}
                        </>
                    )}
                </div>
            )}


            <div className="border-t border-gray-200 my-2"></div>
            <MarketProjectionLayout
                marketId={marketId}
                amount={betAmount}
                direction={selectedOutcome}
            />
            </div>
        </div>
    );
};

export default BuySharesLayout;
