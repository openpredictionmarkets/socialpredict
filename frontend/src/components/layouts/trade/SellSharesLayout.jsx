import React, { useState, useEffect } from 'react';
import { SharesBadge, SaleInputAmount, ConfirmSaleButton } from '../../buttons/trade/SellButtons';
import { fetchUserShares, submitSale } from './TradeUtils'
import MarketProjectionLayout from '../marketprojection/MarketProjectionLayout';

const SellSharesLayout = ({ marketId, token, onTransactionSuccess, currentProbability, totalYes, totalNo }) => {
    const [shares, setShares] = useState({ NoSharesOwned: 0, YesSharesOwned: 0 });
    const [sellAmount, setSellAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);

    useEffect(() => {
        fetchUserShares(marketId, token)
            .then(data => {
                setShares(data);
                // Automatically select an outcome and set initial sell amount
                if (data.NoSharesOwned > 0 && data.YesSharesOwned === 0) {
                    setSelectedOutcome('NO');
                    setSellAmount(data.NoSharesOwned);
                } else if (data.YesSharesOwned > 0 && data.NoSharesOwned === 0) {
                    setSelectedOutcome('YES');
                    setSellAmount(data.YesSharesOwned);
                }
            })
            .catch(error => {
                alert(`Error fetching shares: ${error.message}`);
                console.error('Error fetching shares:', error);
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
                console.log('Sale response:', data);
                onTransactionSuccess();
            }, (error) => {
                alert(`Sale failed: ${error.message}`);
                console.error('Sale error:', error);
            }
        );
    };


    return (
        <div className="p-6 bg-blue-900 rounded-lg text-white">
            <h2 className="text-xl mb-4">Shares Owned</h2>
            {shares.NoSharesOwned < 1 && shares.YesSharesOwned < 1 ? (
                <div className="text-center">
                    <p>No Shares Owned In This Market</p>
                </div>
            ) : (
                <>
                    <div className="flex justify-center space-x-4 mb-4">
                        {shares.NoSharesOwned > 0 &&
                            <SharesBadge type="NO" count={shares.NoSharesOwned} />}
                        {shares.YesSharesOwned > 0 &&
                            <SharesBadge type="YES" count={shares.YesSharesOwned} />}
                    </div>
                    <div className="border-t border-gray-200 my-2"></div>
                    <div className="flex items-center space-x-4 mb-4">
                        <h2 className="text-xl">Sale Amount</h2>
                        <SaleInputAmount
                            value={sellAmount}
                            onChange={handleSellAmountChange}
                            max={selectedOutcome === 'NO' ? shares.NoSharesOwned : shares.YesSharesOwned}
                        />
                    </div>
                    <div className="flex justify-center">
                        {shares.NoSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('NO')} selectedDirection="NO">Sell NO</ConfirmSaleButton>}
                        {shares.YesSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('YES')} selectedDirection="YES">Sell YES</ConfirmSaleButton>
                        }
                    </div>
                    <div className="border-t border-gray-200 my-2"></div>
                    <MarketProjectionLayout
                        currentProbability={currentProbability}
                        totalYes={totalYes}
                        totalNo={totalNo}
                        addedYes={selectedOutcome === 'YES' ? sellAmount : 0}
                        addedNo={selectedOutcome === 'NO' ? sellAmount : 0}
                    />                    
                </>
            )}
        </div>
    );
};

export default SellSharesLayout;
