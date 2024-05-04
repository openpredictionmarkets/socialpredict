import { API_URL } from '../../../config';
import { useEffect, useState } from 'react';

export const submitBet = () => {

    const [token, setToken] = useState(null);

    useEffect(() => {
        setToken(localStorage.getItem('token'));
    }, []);

    if (!token) {
        alert('Please log in to place a bet.');
        return;
    }

    const betData = {
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