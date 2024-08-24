import { API_URL } from '../../../config';
import React, { useState, useEffect } from 'react';

const MarketProjectionLayout = ({ currentProbability, totalYes, totalNo, addedYes = 0, addedNo = 0 }) => {
    const [setupData, setSetupData] = useState(null);


    useEffect(() => {
        fetch('/api/v0/setup')
            .then(response => response.json())
            .then(data => setSetupData(data))
            .catch(error => console.error('Error fetching setup data:', error));
    }, []);

    if (!setupData) {
        return <div>Loading...</div>;
    }

    // Extract necessary values from setupData
    const P_initial = setupData.MarketCreation.InitialMarketProbability;
    const I_initial = setupData.MarketCreation.InitialMarketSubsidization;

    // Calculate the next probability
    const nextProbability = (P_initial * I_initial + totalYes + addedYes) / 
                            (I_initial + totalYes + addedYes + totalNo + addedNo);

    return (
        <div className="market-projection-layout">
            <div className="projection-details">
                <p>New Market Probability: {nextProbability.toFixed(4)}</p>
            </div>
        </div>
    );
};

export default MarketProjectionLayout;
