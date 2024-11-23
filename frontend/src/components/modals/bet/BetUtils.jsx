import { API_URL } from '../../../config.js';

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
                return response.json().then(err => {
                    // Construct a new error object and throw it
                    throw new Error(`Error placing bet: ${err.error || 'Unknown error'}`);
                });
            }
            return response.json();
        })
        .then(data => {
            console.log('Success:', data);
            onSuccess(data);  // Handle success outside this utility function
        })
        .catch(error => {
            console.error('Error:', error);
            alert(error.message);  // Use error.message to display the custom error message
            onError(error);  // Handle error outside this utility function
        });
};
