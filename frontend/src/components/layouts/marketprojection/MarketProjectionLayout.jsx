import { API_URL } from '../../../config';
import React, { useState } from 'react';
import { SiteButtonSmall } from '../../buttons/SiteButtons';

const MarketProjectionLayout = ({ marketId, amount, direction }) => {
    const [projectionData, setProjectionData] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const fetchProjectionData = async () => {
        setLoading(true);
        setError(null);

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

    const canProject = Boolean(amount) && Boolean(direction);

    return (
        <div className="market-projection-layout">
            <div className="projection-details">
                {!canProject && (
                    <p className="text-sm text-gray-300 mb-2">
                        Select amount and outcome to see the projected probability after this trade.
                    </p>
                )}
                <SiteButtonSmall
                    className="mb-4"
                    onClick={fetchProjectionData}
                    disabled={!canProject || loading}
                >
                    {loading ? 'Updating...' : 'Update Projection'}
                </SiteButtonSmall>
                {error && <div className="error-message">Error: {error}</div>}
                {projectionData && typeof projectionData.projectedProbability === 'number' && (
                    <p>New Market Probability: {projectionData.projectedProbability.toFixed(4)}</p>
                )}
            </div>
        </div>
    );
};

export default MarketProjectionLayout;
