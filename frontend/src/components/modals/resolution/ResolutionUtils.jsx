
export const resolveMarket = () => {
    const resolutionData = {
        outcome: selectedResolution,
        percentage: resolutionPercentage,
    };

    const requestOptions = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(resolutionData),
    };

    fetch(`${API_URL}/api/v0/resolve/${marketId}`, requestOptions)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            console.log('Market resolved successfully:', data);
            window.location.reload(); // Reload or redirect
        })
        .catch(error => {
            console.error('Error resolving market:', error);
        });

    setShowResolveModal(false);
};