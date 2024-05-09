import React, { useState, useEffect } from 'react';
// import { BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/BetButtons';
import { SharesBadge, SaleInputAmount, ConfirmSaleButton } from '../../buttons/SellButtons';
import { fetchUserShares, submitSale } from './TradeUtils'

const SellSharesLayout = ({ marketId, token, onSaleSuccess }) => {
    const [shares, setShares] = useState({ NoSharesOwned: 0, YesSharesOwned: 0 });
    const [sellAmount, setSellAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);

    useEffect(() => {
        fetchUserShares(marketId, token)
            .then(data => {
                setShares(data);
                // Automatically select an outcome if only one type of share is owned
                if (data.NoSharesOwned > 0 && data.YesSharesOwned === 0) {
                    setSelectedOutcome('NO');
                } else if (data.YesSharesOwned > 0 && data.NoSharesOwned === 0) {
                    setSelectedOutcome('YES');
                }
            })
            .catch(error => {
                alert(`Error fetching shares: ${error.message}`);
                console.error('Error fetching shares:', error);
            });
    }, [marketId, token]);

    const handleSellAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10);
        setSellAmount(newAmount >= 0 ? newAmount : '');
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
                onSaleSuccess();
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
                        <SaleInputAmount value={sellAmount} onChange={handleSellAmountChange} />
                    </div>
                    <div className="flex justify-center">
                        {shares.NoSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('NO')} selectedDirection="NO">Sell NO</ConfirmSaleButton>}
                        {shares.YesSharesOwned > 0 &&
                            <ConfirmSaleButton onClick={() => handleSaleSubmission('YES')} selectedDirection="YES">Sell YES</ConfirmSaleButton>}
                    </div>
                </>
            )}
        </div>
    );
};

export default SellSharesLayout;
