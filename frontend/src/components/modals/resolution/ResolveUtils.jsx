import { API_URL } from '../../../config';

export const resolveMarket = (marketId, token, selectedResolution) => {
    const resolutionData = {
        outcome: selectedResolution,
    };

    const requestOptions = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(resolutionData),
    };

    // Returning fetch promise to allow handling of the response in the component
    return fetch(`${API_URL}/api/v0/resolve/${marketId}`, requestOptions)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            console.log('Market resolved successfully:', data);
            return data;
        })
        .catch(error => {
            console.error('Error resolving market:', error);
            throw error;
        });
};
