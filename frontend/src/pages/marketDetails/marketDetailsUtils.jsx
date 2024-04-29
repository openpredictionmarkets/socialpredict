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
                setDetails(data);  // Assuming data is structured as in your example
            } catch (error) {
                console.error('Error fetching market data:', error);
            }
        };
        fetchData();
    }, [marketId]);

    return details;
};
