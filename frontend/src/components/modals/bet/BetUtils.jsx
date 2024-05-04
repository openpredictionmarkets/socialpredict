// import API_URL from your config
import { API_URL } from '../../../config';

export const submitBet = (betData, token, onSuccess, onError) => {


    if (!token) {
        alert('Please log in to place a bet.');
        return;
    }

    console.log('Sending bet data:', betData);

    fetch(`${API_URL}/api/v0/bet`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(betData),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Success:', data);
        onSuccess(data);  // Handle success outside this utility function
    })
    .catch(error => {
        console.error('Error:', error);
        onError(error);  // Handle error outside this utility function
    });
};
