import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { API_URL } from '../../config';

export const fetchMarketDataHook = () => {
    const [details, setDetails] = useState(null);
    const { marketId } = useParams();

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/markets/${marketId}`);
                const data = await response.json();
                setDetails(data);
            } catch (error) {
                console.error('Error fetching market data:', error);
            }
        };
        fetchData();
    }, [marketId]);

    return details;
};

export const useFetchMarketData = () => {
    const [details, setDetails] = useState(null);
    const { marketId } = useParams();
    const [triggerRefresh, setTriggerRefresh] = useState(false);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch(`${API_URL}/api/v0/markets/${marketId}`);
                const data = await response.json();
                setDetails(data);
            } catch (error) {
                console.error('Error fetching market data:', error);
            }
        };

        fetchData();
    }, [marketId, triggerRefresh]);

    const refetchData = () => {
        setTriggerRefresh(prev => !prev);
    };

    return { details, refetchData };
};

export const calculateCurrentProbability = (details) => {
    if (!details || !details.probabilityChanges) return 0;

    const currentProbability = details.probabilityChanges.length > 0
        ? details.probabilityChanges[details.probabilityChanges.length - 1].probability
        : details.market.initialProbability;

    return parseFloat(currentProbability.toFixed(3));
};







//const roundedProbability = parseFloat(newCurrentProbability.toFixed(3));

        // Append the current time with the last known probability, converted to Unix time
//        const [currentProbability, setCurrentProbability] = useState(null);
//        const currentTimeStamp = new Date().getTime();
//        chartData.push({ time: currentTimeStamp, P: roundedProbability });
