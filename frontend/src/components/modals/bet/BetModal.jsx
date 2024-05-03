import React, { useState } from 'react';

const BetModalButton = ({ marketId, token }) => {
    const [showBetModal, setShowBetModal] = useState(false);
    const [betAmount, setBetAmount] = useState(20);
    const [selectedOutcome, setSelectedOutcome] = useState(null);

    const toggleBetModal = () => setShowBetModal(prev => !prev);

    const handleBetAmountChange = (event) => {
        setBetAmount(parseFloat(event.target.value) || 0);
    };

    const submitBet = () => {
        if (!token) {
            alert('Please log in to place a bet.');
            return;
        }

        const betData = {
            username: localStorage.getItem('username'),  // Assuming username is stored in localStorage
            marketId,
            amount: betAmount,
            outcome: selectedOutcome,
        };

        console.log('Sending bet data:', betData);

        fetch(`${API_URL}/api/v0/bet`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify(betData),
        }).then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        }).then(data => {
            console.log('Success:', data);
            alert(`Bet placed successfully! Bet ID: ${data.id}`);
            setShowBetModal(false);
        }).catch(error => {
            console.error('Error:', error);
            alert(`Error placing bet: ${error.message}`);
        });
    };

    return (
        <div>
            <button onClick={toggleBetModal}>Place Bet</button>
            {showBetModal && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
                    <div className="bet-modal relative bg-blue-900 p-6 rounded-lg text-white m-6 mx-auto" style={{ width: '350px' }}>
                        <h2 className="text-xl mb-4">Place Your Bet</h2>
                        <input type="number" value={betAmount} onChange={handleBetAmountChange} className="text-black mb-4"/>
                        <button onClick={() => setSelectedOutcome('YES')} className="mr-4">Bet YES</button>
                        <button onClick={() => setSelectedOutcome('NO')}>Bet NO</button>
                        <button onClick={submitBet} className="mt-4">Confirm Bet</button>
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
