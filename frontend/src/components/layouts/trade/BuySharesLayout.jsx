import React, { useState, useEffect } from 'react';
import { BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/trade/BetButtons';
import MarketProjectionLayout from '../marketprojection/MarketProjectionLayout';
import { submitBet } from './TradeUtils';
import { useMarketLabels } from '../../../hooks/useMarketLabels';
import { fetchTradingFees } from '../../../api/tradeApi';
import { USER_CREDIT_REFRESH_EVENT } from '../../utils/userFinanceTools/FetchUserCredit';


const BuySharesLayout = ({ marketId, market, token, onTransactionSuccess }) => {
    const [betAmount, setBetAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);
    const [feeData, setFeeData] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [showTerms, setShowTerms] = useState(false);
    
    // Get custom labels for this market
    const { yesLabel, noLabel } = useMarketLabels(market);
    const showFeeSection = !isLoading && feeData;
    const hasVisibleFee = showFeeSection && (feeData.initialBetFee > 0 || feeData.buySharesFee > 0);

    useEffect(() => {
        const fetchFeeData = async () => {
            if (!token) {
                setIsLoading(false);
                return;
            }
            try {
                setFeeData(await fetchTradingFees({ token }));
            } catch {
                setFeeData(null);
            } finally {
                setIsLoading(false);
            }
        };

        fetchFeeData();
    }, [token]);


    const handleBetAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10);
        setShowTerms(false);
        setBetAmount(newAmount >= 0 ? newAmount : '');
    };

    const handleOutcomeSelection = (outcome) => {
        setSelectedOutcome(outcome);
        setShowTerms(false);
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
            window.dispatchEvent(new Event(USER_CREDIT_REFRESH_EVENT));
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
                    onClick={() => handleOutcomeSelection('NO')} 
                    label={noLabel}
                />
                <BetYesButton 
                    onClick={() => handleOutcomeSelection('YES')} 
                    label={yesLabel}
                />
            </div>
            <div className="border-t border-gray-200 my-2"></div>
            <div className="flex items-center space-x-4 mb-4">
                <h2 className="text-xl">Amount</h2>
                <BetInputAmount value={betAmount} onChange={handleBetAmountChange} />
            </div>
            <ConfirmBetButton onClick={handleBetSubmission} selectedDirection={selectedOutcome} yesLabel={yesLabel} noLabel={noLabel} />
            <button
                type="button"
                onClick={() => setShowTerms((current) => !current)}
                className="mt-2 w-full rounded border border-blue-200/60 px-4 py-2 text-sm font-semibold text-blue-50 hover:bg-white/10"
            >
                Terms
            </button>
            {showTerms && (
                <div className="mt-3 border-t border-gray-200 pt-3">
                    {hasVisibleFee && (
                        <div className="mb-4">
                            {feeData.initialBetFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Initial Trade Fee: {feeData.initialBetFee}
                                    <span className="block">Does not apply if already traded on this market.</span>
                                </p>
                            )}
                            {feeData.buySharesFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Trading Fee (Buying Share): {feeData.buySharesFee}
                                </p>
                            )}
                        </div>
                    )}
                    <MarketProjectionLayout
                        marketId={marketId}
                        amount={betAmount}
                        direction={selectedOutcome}
                    />
                </div>
            )}
        </div>
    );
};

export default BuySharesLayout;
