import { API_URL } from '../../../config';
import React, { useState, useEffect } from 'react';

const MarketProjectionLayout = ({ marketId, amount, direction }) => {
    const [projectionData, setProjectionData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchProjectionData = async () => {
            try {
                const response = await fetch(`${API_URL}/v0/marketprojection/${marketId}/${amount}/${direction}/`);
                if (!response.ok) {
                    throw new Error(`Error fetching data: ${response.statusText}`);
                }
                const data = await response.json();
                setProjectionData(data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchProjectionData();
    }, [marketID, amount, direction]);

    if (loading) {
        return <div>Loading...</div>;
    }

    if (error) {
        return <div>Error: {error}</div>;
    }

    if (!projectionData) {
        return <div>No data available</div>;
    }
    
    return (
        <div className="market-projection-layout">
            <div className="projection-details">
                <p>New Market Probability: {projectionData.projectedprobability.toFixed(4)}</p>
            </div>
        </div>
    );
};

export default MarketProjectionLayout;
