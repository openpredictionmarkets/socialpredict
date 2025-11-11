import { API_URL } from '../../../config';
import React, { useState } from 'react';

const MarketProjectionLayout = ({ marketId, amount, direction }) => {
    const [projectionData, setProjectionData] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const fetchProjectionData = async () => {
        if (!amount || !direction) {
            setError('Amount and direction are required to fetch the market projection.');
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const response = await fetch(`${API_URL}/v0/marketprojection/${marketId}/${amount}/${direction}/`);
            if (!response.ok) {
                throw new Error(`Error fetching data: ${response.statusText}`);
            }

            // Log the raw text of the response
            const responseText = await response.text();
            console.log('Response text:', responseText);

            // Try parsing the response as JSON
            const data = JSON.parse(responseText);

            // Log the parsed JSON data
            console.log('Parsed JSON data:', data);

            setProjectionData(data);
        } catch (err) {
            setError(err.message);
            console.error('Error fetching market projection:', err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="market-projection-layout">
            <div className="projection-details">
                <button
                    className="btn btn-primary mb-4"
                    onClick={fetchProjectionData}
                    disabled={!amount || !direction || loading}
                >
                    {loading ? 'Updating...' : 'Update Projection'}
                </button>
                {error && <div className="error-message">Error: {error}</div>}
                {projectionData && (
                    <p>New Market Probability: {projectionData.projectedprobability.toFixed(2)}</p>
                )}
                {!projectionData && !error && !loading && (
                    <p>Click "Update Projection" to see the new market probability after this trade.</p>
                )}
            </div>
        </div>
    );
};

export default MarketProjectionLayout;
