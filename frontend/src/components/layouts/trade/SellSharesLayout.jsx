import React, { useState, useEffect } from 'react';
import { SharesBadge, SaleInputAmount, ConfirmSaleButton } from '../../buttons/trade/SellButtons';
import { fetchUserShares, submitSale } from './TradeUtils';
import { useMarketLabels } from '../../../hooks/useMarketLabels';
import { API_URL } from '../../../config';

const SellSharesLayout = ({ marketId, market, token, onTransactionSuccess }) => {
    const [shares, setShares] = useState({ NoSharesOwned: 0, YesSharesOwned: 0 });
    const [sellAmount, setSellAmount] = useState(1);
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
            } catch {
                setFeeData(null);
            } finally {
                setIsLoading(false);
            }
        };

        fetchFeeData();
    }, [token]);

    useEffect(() => {
        fetchUserShares(marketId, token)
            .then(data => {
                // Always get an object, never an array
                let sharesObj;
                if (Array.isArray(data)) {
                    sharesObj = data[0] || { noSharesOwned: 0, yesSharesOwned: 0, value: 0 };
                } else if (typeof data === 'object' && data !== null) {
                    sharesObj = data;
                } else {
                    sharesObj = { noSharesOwned: 0, yesSharesOwned: 0, value: 0 };
                }
                setShares(sharesObj);

                // Set outcome and amount based on shares
                if (sharesObj.noSharesOwned > 0 && sharesObj.yesSharesOwned === 0) {
                    setSelectedOutcome('NO');
                    setSellAmount(sharesObj.noSharesOwned);
                } else if (sharesObj.yesSharesOwned > 0 && sharesObj.noSharesOwned === 0) {
                    setSelectedOutcome('YES');
                    setSellAmount(sharesObj.yesSharesOwned);
                }
            })
            .catch(error => {
                alert(`Error fetching shares: ${error.message}`);
                // Optionally, reset to default state on error
                setShares({ noSharesOwned: 0, yesSharesOwned: 0, value: 0 });
            });
    }, [marketId, token]);


    const handleSellAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10) || 0; // Ensure it defaults to 0 if conversion fails
        // Check the selected outcome and compare the new amount with the owned shares
        if (selectedOutcome === 'NO') {
            if (newAmount > shares.NoSharesOwned) {
                setSellAmount(shares.NoSharesOwned); // Set to max shares if over the limit
            } else if (newAmount >= 0) {
                setSellAmount(newAmount); // Only set if it's a non-negative number
            }
        } else if (selectedOutcome === 'YES') {
            if (newAmount > shares.YesSharesOwned) {
                setSellAmount(shares.YesSharesOwned); // Set to max shares if over the limit
            } else if (newAmount >= 0) {
                setSellAmount(newAmount); // Only set if it's a non-negative number
            }
        }
    };

    const handleSaleSubmission = () => {
        const saleData = {
            marketId: marketId,
            outcome: selectedOutcome,
            amount: sellAmount,
        };

        submitSale(saleData, token, (data) => {
                alert(`Sale successful! Sale ID: ${data.id}`);
                onTransactionSuccess();
            }, (error) => {
                alert(`Sale failed: ${error.message}`);
            }
        );
    };


    return (
        <div className="p-6 bg-blue-900 rounded-lg text-white">
            <h2 className="text-xl mb-4">Shares Owned</h2>
            {shares.noSharesOwned < 1 && shares.yesSharesOwned < 1 ? (
                <div className="text-center">
                    <p>No Shares Owned In This Market</p>
                </div>
            ) : (
                <>
                    <div className="flex justify-center space-x-4 mb-4">
                        {shares.noSharesOwned > 0 &&
                            <SharesBadge type="NO" count={shares.noSharesOwned} label={noLabel} />}
                        {shares.yesSharesOwned > 0 &&
                            <SharesBadge type="YES" count={shares.yesSharesOwned} label={yesLabel} />}
                    </div>
                    {(shares.noSharesOwned > 0 || shares.yesSharesOwned > 0) && (
                        <div className="text-center text-lg mt-2">
                            <span className="font-bold">Position Value: </span>
                            <span className="text-green-300">{shares.value}</span>
                        </div>
                    )}
                    <div className="border-t border-gray-200 my-2"></div>
                    <div className="flex items-center space-x-4 mb-4">
                        <h2 className="text-xl">Sale Amount</h2>
                        <SaleInputAmount
                            value={sellAmount}
                            onChange={handleSellAmountChange}
                            max={selectedOutcome === 'NO' ? shares.noSharesOwned : shares.yesSharesOwned}
                        />
                    </div>
                    <div className="flex justify-center">
                        {shares.noSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('NO')} selectedDirection="NO">Sell NO</ConfirmSaleButton>}
                        {shares.yesSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('YES')} selectedDirection="YES">Sell YES</ConfirmSaleButton>}
                    </div>
                </>
            )}

            <div className="border-t border-gray-200 my-2"></div>

            {!isLoading && feeData && (
                <div className="mb-4">
                    {feeData.InitialBetFee === 0 && feeData.SellSharesFee === 0 ? (
                        <p className="text-sm text-gray-300">No fees</p>
                    ) : (
                        <>
                            {feeData.InitialBetFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Initial Trade Fee: {feeData.InitialBetFee}
                                    <span className="block">Does not apply if already traded on this market.</span>
                                </p>
                            )}
                            {feeData.SellSharesFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Trading Fee (Selling Share): {feeData.SellSharesFee}
                                </p>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    );


};

export default SellSharesLayout;
