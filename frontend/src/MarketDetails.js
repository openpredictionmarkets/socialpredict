import React, { useState, useEffect, useContext, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import UserContext from './UserContext';
import './MarketDetails.css';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';


// chart stuff

// format the date for the axis on the chart

const formatDateForAxis = (unixTime) => {
    const date = new Date(unixTime);
    const day = date.getDate();
    const month = date.toLocaleString('en-US', { month: 'short' }); // 'short' gives the three-letter abbreviation
    return `${day}-${month.toUpperCase()}`; // Formats the date as DD-MMM
};

// custom tool tip on the chart

const CustomTooltip = ({ active, payload }) => {
    if (active && payload && payload.length) {
        // Convert Unix time to ISO format
        const date = new Date(payload[0].payload.time);
        const isoFormat = date.toISOString();

        // Round the probability to 3 decimal places
        const probability = payload[0].value.toFixed(3);

        return (
            <div className="custom-tooltip">
                <p>{`Time: ${isoFormat}`}</p>
                <p>{`Probability: ${probability}`}</p>
            </div>
        );
    }

    return null;
};

// format the time for the main display grid, showing resolution time
function formatDateTimeForGrid(dateTimeString) {
    const date = new Date(dateTimeString);

    // Check if there are any seconds
    if (date.getSeconds() > 0) {
        // Add one minute
        date.setMinutes(date.getMinutes() + 1);
        // Reset seconds to zero
        date.setSeconds(0);
    }

    // Extracting date components
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0'); // Months are 0-based
    const day = date.getDate().toString().padStart(2, '0');


    // Getting timezone
    const timezone = /\(([^)]+)\)$/.exec(date.toString())[1];

    // Convert 24-hour time to 12-hour time and determine AM/PM
    let hour = date.getHours();
    const amPm = hour >= 12 ? 'PM' : 'AM';
    hour = hour % 12;
    hour = hour ? hour : 12; // the hour '0' should be '12'
    const formattedHour = hour.toString().padStart(2, '0');

    const minute = date.getMinutes().toString().padStart(2, '0');

    // Format to YYYY.MM.DD and append time with AM/PM
    return `${year}.${month}.${day} ${formattedHour}:${minute} ${amPm}`;
};

// overall marketDetails function

function MarketDetails() {

    // User Context
    const { username } = useContext(UserContext);
    const isLoggedIn = username !== null;
    const token = localStorage.getItem('token');

    // Market Logic
    const [market, setMarket] = useState(null);
    const [currentProbability, setCurrentProbability] = useState(null);
    const [chartData, setChartData] = useState([]);
    const { marketId } = useParams();
    const [numUsers, setNumUsers] = useState(0);
    const [totalVolume, setTotalVolume] = useState(0);

    // Bet logic
    const [betAmount, setBetAmount] = useState(20); // Default bet amount
    const [showBetModal, setShowBetModal] = useState(false); // State to control modal visibility
    const [selectedOutcome, setSelectedOutcome] = useState(null); // State to store selected outcome

    // Resolving Market Control
    const [showResolveModal, setShowResolveModal] = useState(false);
    const openResolveModal = () => {
        setShowResolveModal(prev => !prev);
    };
    const [selectedResolution, setSelectedResolution] = useState(null);
    const [resolutionPercentage, setResolutionPercentage] = useState(0);

    const resolveMarket = () => {
        const resolutionData = {
            outcome: selectedResolution
            // Removed marketId from here, as it's now part of the URL
        };

        const requestOptions = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(resolutionData)
        };

        fetch(`https://brierfoxforecast.ngrok.app/api/v0/resolve/${marketId}`, requestOptions)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                console.log('Market resolved successfully:', data);
                // Handle success, maybe refresh the market details or redirect
                window.location.reload();
            })
            .catch(error => {
                console.error('Error resolving market:', error);
                // Handle errors, maybe show a notification to the user
            });

        setShowResolveModal(false); // Close the modal after submission
    };

    // get the Market Data, information from the endpoint.
    const fetchMarketData = useCallback(() => {
        fetch(`https://brierfoxforecast.ngrok.app/api/v0/markets/${marketId}`)
            .then(response => response.json())
            .then(data => {
                setMarket(data.market);
                setNumUsers(data.numUsers); // Assuming `numUsers` is part of the response
                setTotalVolume(data.totalVolume); // Assuming `totalVolume` is part of the response

                // Extract the last probability value
                const probabilityChanges = data.probabilityChanges;
                const newCurrentProbability = probabilityChanges.length > 0
                    ? probabilityChanges[probabilityChanges.length - 1].probability
                    : data.market.initialProbability;

                // Round to 3 decimal places
                const roundedProbability = parseFloat(newCurrentProbability.toFixed(3));

                // Set data to display in website
                setCurrentProbability(roundedProbability);

                // Convert the timestamps to Unix time (milliseconds since the Unix epoch)
                let chartData = data.probabilityChanges.map(change => ({
                    time: new Date(change.timestamp).getTime(),
                    P: change.probability

                }));

                // Append the current time with the last known probability, converted to Unix time
                const currentTimeStamp = new Date().getTime();
                chartData.push({ time: currentTimeStamp, P: roundedProbability });

                setChartData(chartData);
            })
            .catch(error => console.error('Error:', error));
    }, [marketId]);


    useEffect(() => {
        fetchMarketData();
    }, [fetchMarketData]);


    // condition if there are no markets
    if (!market) {
        return <div>Loading...</div>;
    }

    // Betting Stuff, open the modal, handle the amount entered in the box, submit bet

    const openBetModal = (outcome) => {
        setSelectedOutcome(outcome);
        setShowBetModal(true);
    };

    const handleBetAmountChange = (event) => {
        const newAmount = parseFloat(event.target.value);
        setBetAmount(newAmount);
    };

    const submitBet = () => {
        setShowBetModal(false);
        placeBet(selectedOutcome, betAmount).then(() => {
            fetchMarketData(); // Refresh market data after placing the bet
        });
    };

    // logic for placing bets
    const placeBet = (outcome) => {

        return new Promise((resolve, reject) => {
            if (!isLoggedIn) {
                alert("Please log in to place a bet.");
                reject(new Error("Not logged in"));
                return;
            }

            // Retrieve the JWT token from localStorage
            const token = localStorage.getItem('token');
            if (!token) {
                alert("Authentication token not found. Please log in again.");
                return reject(new Error("Token not found"));
            }

            const betData = {
                username: username,
                marketId: parseInt(marketId, 10),
                amount: parseFloat(betAmount), // Ensure amount is a float
                outcome: outcome
            };

            console.log("Sending bet data:", betData);

            fetch('https://brierfoxforecast.ngrok.app/api/v0/bet', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(betData)
            })
            // ensure we can handle non-json error responses for complicated problems
            .then(response => {
                if (!response.ok) {
                    // Handle non-JSON error response
                    if (response.headers.get("content-type") !== "application/json") {
                        throw new Error(`Non-JSON error from server: ${response.statusText}`);
                    }
                    return response.json().then(errorBody => {
                        throw new Error(`Error ${response.status}: ${errorBody.message}`);
                    });
                }
                return response.json();
            })
            .then(data => {
                console.log('Success:', data);
                alert(`Bet placed successfully! Bet ID: ${data.id}`);
                resolve(data); // Resolve the promise with the data
            })
            .catch(error => {
                console.error('Error:', error);
                alert(`Error placing bet: ${error.message}`);
                reject(error); // Reject the promise on error
            });
        });
    };


    return (
        <div>
            <h3>{market.questionTitle}</h3>
            <table className="skinny-table">
                <tbody>
                    <tr>
                        <td>
                        <div className="nav-link">
                            <Link to={`/user/${market.creatorUsername}`} className="nav-link">
                            😀 @{market.creatorUsername}
                            </Link>
                        </div>
                        </td>
                        <td>👤 {numUsers}</td>
                        <td>📊 {totalVolume.toFixed(2)}</td>
                        <td>💬 0</td>
                        <td>
                            {market.isResolved ? (
                                <span>
                                    RESOLVED: {market.resolutionResult}
                                    <p>
                                    @ {formatDateTimeForGrid(market.finalResolutionDateTime).toLocaleString()}
                                    </p>
                                </span>
                            ) : (
                                <span>
                                    Ends: 📅 {formatDateTimeForGrid(market.resolutionDateTime).toLocaleString()}
                                </span>
                            )}
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="5" className="probability">
                            {market.isResolved ? (
                                <span>
                                <p style={{ textAlign: 'left' }}>Final Probability:</p>
                                <h2 style={{ textAlign: 'left' }}>{currentProbability}</h2>
                                </span>
                            ) : (
                                <span>
                                <p style={{ textAlign: 'left' }}>Current Probability:</p>
                                <h2 style={{ textAlign: 'left' }}>{currentProbability}</h2>
                                </span>
                            )}
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="5">
                        <LineChart width={1050} height={300} data={chartData}>
                            <XAxis
                                dataKey="time"
                                type="number"
                                domain={['auto', 'auto']}
                                tickFormatter={formatDateForAxis}
                            />
                            <YAxis domain={[0, 1]} />
                            <CartesianGrid strokeDasharray="3 3" />
                            <Tooltip content={<CustomTooltip />} />
                            <Legend />
                            <Line type="linear" dataKey="P" stroke="#8884d8" />
                        </LineChart>
                        </td>
                    </tr>
                </tbody>
            </table>

            {isLoggedIn && !market.isResolved && (
            <div className="bet-decision-buttons">
                <button className="bet-decision-button bet-yes-button" onClick={() => openBetModal("YES")}>YES</button>
                <button className="bet-decision-button bet-no-button" onClick={() => openBetModal("NO")}>NO</button>
            </div>
            )}

            {showBetModal && (
                <div className="bet-decision-buttons">
                    <input className="bet-input" type="number" value={betAmount} onChange={handleBetAmountChange} />
                    <button className={`confirm-button ${selectedOutcome}-button`} onClick={submitBet}>CONFIRM</button>
                    <button className={`cancel-button`} onClick={() => setShowBetModal(false)}>CANCEL</button>
                </div>
            )}


            {isLoggedIn && !market.isResolved && username === market.creatorUsername && (
                <button className="resolve-button"
                onClick={openResolveModal}
                >
                    Resolve
                </button>
            )}

            {showResolveModal && (
                <div className="resolve-modal">
                    <div className="resolve-container">
                        <button
                            className={`resolve-decision-button ${selectedResolution === 'YES' ? 'yes-selected' : ''}`}
                            onClick={() => setSelectedResolution('YES')}
                        >
                            YES
                        </button>
                        <button
                            className={`resolve-decision-button ${selectedResolution === 'NO' ? 'no-selected' : ''}`}
                            onClick={() => setSelectedResolution('NO')}
                        >
                            NO
                        </button>
                        <button
                            className={`resolve-decision-button ${selectedResolution === 'N/A' ? 'na-selected' : ''}`}
                            onClick={() => setSelectedResolution('N/A')}
                        >
                            N/A
                        </button>
                        <input
                type="number"
                className={`percent-resolution-input ${selectedResolution === '%' ? 'percent-input' : ''}`}
                value={resolutionPercentage}
                onChange={(e) => {
                    let newValue = parseInt(e.target.value, 10);
                    if (newValue < 1) newValue = 1;
                    if (newValue > 99) newValue = 99;
                    setResolutionPercentage(newValue);
                    setSelectedResolution('%');
                }}
                min="1"
                max="99"
            />
            <button
                className={`resolve-decision-button ${selectedResolution === '%' ? 'percent-selected' : ''}`}
                onClick={() => setSelectedResolution('%')}
            >
                %
            </button>
                    </div>
                    <button
                        className="resolve-confirm-button"
                        onClick={() => resolveMarket()}
                    >
                        CONFIRM RESOLUTION
                    </button>
                </div>
            )}




        </div>
    );
}

export default MarketDetails;