import { API_URL } from '../../../config';
import { parseApiError } from '../../../utils/apiError';

export const submitBet = (betData, token, onSuccess, onError) => {

    if (!token) {
        alert('Please log in to place a bet.');
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
        alert(error.message);
        onError(error);
    });
};
