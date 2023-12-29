import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import './Markets.css';

function Markets() {
    const [marketsData, setMarketsData] = useState([]);

    useEffect(() => {
        // Fetch data from your API
        fetch('https://brierfoxforecast.ngrok.app/api/v0/markets') // Adjust the URL to your API endpoint
            .then(response => response.json())
            .then(data => {
                setMarketsData(data); // Set the fetched markets data
            })
            .catch(error => console.error('Error fetching market data:', error));
    }, []); // Empty dependency array means this effect runs once on mount

    // Render the markets or a loading message
    return (
        <div className="App">
            <div className="Center-content">
            <div className="Center-content-header"><h1>Markets</h1></div>
            <div className="Center-content-table">
                {marketsData.length > 0 ? (
                    <table>
                        <tbody>
                            {marketsData.map(market => (
                                <tr key={market.id}>
                                    <td>⬆️⬇️</td>
                                    <td>{market.initialProbability}</td>
                                    <td><Link to={`/markets/${market.id}`}>{market.questionTitle}</Link></td>
                                    <td>📅 {new Date(market.resolutionDateTime).toLocaleDateString()}</td>
                                    <td>😀 admin</td> {/* Placeholder for actual data */}
                                    <td>👤 20</td>
                                    <td>📊 1.5k</td>
                                    <td>💬 12</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                ) : (
                    <div>Loading markets...None may be available.</div>
                )}
            </div>
            </div>
        </div>
    );
}

export default Markets;