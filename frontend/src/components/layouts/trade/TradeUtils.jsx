import { API_URL } from '../../../config';
import { parseApiError } from '../../../utils/apiError';

export const submitBet = (betData, token, onSuccess, onError) => {

    if (!token) {
        alert('Please log in to place a bet.');
        return;
    }

    // Validate bet data before sending
    if (!betData.marketId || !betData.amount || !betData.outcome) {
        onError(new Error('Missing required bet data (marketId, amount, outcome)'));
        return;
    }

    if (betData.amount < 1) {
        onError(new Error('Bet amount must be at least 1'));
        return;
    }

    if (betData.outcome !== 'YES' && betData.outcome !== 'NO') {
        onError(new Error('Bet outcome must be YES or NO'));
        return;
    }

    console.log('Sending bet data:', betData);

    fetch(`${API_URL}/v0/bet`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(betData),
    })
    .then(async response => {
        if (!response.ok) {
            const msg = await parseApiError(response);
            throw new Error(msg);
        }
        return response.json();
    })
    .then(data => {
        onSuccess(data);
    })
    .catch(error => {
        console.error('Error:', error);
        onError(error);
    });
};


export const fetchUserShares = async (marketId, token) => {
    const response = await fetch(`${API_URL}/v0/userposition/${marketId}`, {
        headers: {
        'Authorization': `Bearer ${token}` // Include the authorization token to validate user
        }
    });
    if (!response.ok) {
        throw new Error('Failed to fetch user shares');
    }
    return response.json();
    };


export const submitSale = (saleData, token, onSuccess, onError) => {
    if (!token) {
        alert('Please log in to sell shares.');
        return;
    }

    // Validate sale data before sending
    if (!saleData.marketId || !saleData.amount || !saleData.outcome) {
        onError(new Error('Missing required sale data (marketId, amount, outcome)'));
        return;
    }

    if (saleData.amount < 1) {
        onError(new Error('Sale amount must be at least 1'));
        return;
    }

    if (saleData.outcome !== 'YES' && saleData.outcome !== 'NO') {
        onError(new Error('Sale outcome must be YES or NO'));
        return;
    }

    console.log('Sending sale data:', saleData);

    fetch(`${API_URL}/v0/sell`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(saleData),
    })
    .then(async response => {
        if (!response.ok) {
            const msg = await parseApiError(response);
            throw new Error(msg);
        }
        return response.json();
    })
    .then(data => {
        onSuccess(data);
    })
    .catch(error => {
        console.error('Error:', error);
        onError(error);
    });
};
