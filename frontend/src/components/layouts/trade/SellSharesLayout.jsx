import React, { useState, useEffect } from 'react';
// import { BetYesButton, BetNoButton, BetInputAmount, ConfirmBetButton } from '../../buttons/BetButtons';
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
            marketId: selectedMarketId,
            outcome: selectedOutcome, // 'YES' or 'NO'
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
            <div className="flex justify-center space-x-4 mb-4">
                {shares.NoSharesOwned > 0 && (
                    <button onClick={() => setSelectedOutcome('NO')} className={selectedOutcome === 'NO' ? 'selected' : ''}>
                        NO: {shares.NoSharesOwned}
                    </button>
                )}
                {shares.YesSharesOwned > 0 && (
                    <button onClick={() => setSelectedOutcome('YES')} className={selectedOutcome === 'YES' ? 'selected' : ''}>
                        YES: {shares.YesSharesOwned}
                    </button>
                )}
            </div>
            <div className="border-t border-gray-200 my-2"></div>
            <div className="flex items-center space-x-4 mb-4">
                <h2 className="text-xl">Sale Amount</h2>
                <input type="number" value={sellAmount} onChange={handleSellAmountChange} />
            </div>
            <button onClick={handleSaleSubmission}>Confirm Sale</button>
        </div>
    );
};

export default SellSharesLayout;
