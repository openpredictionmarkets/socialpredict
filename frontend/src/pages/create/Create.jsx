import { API_URL } from '../../config';
import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import { getEndofDayDateTime } from '../../components/utils/dateTimeTools/FormDateTimeTools';
import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';
import { RegularInput } from '../../components/inputs/InputBar';
import RegularInputBox from '../../components/inputs/InputBox';
import SiteButton from '../../components/buttons/SiteButtons';


function Create() {
    const [questionTitle, setQuestionTitle] = useState('');
    const [description, setDescription] = useState('');
    const [resolutionDateTime, setResolutionDateTime] = useState(getEndofDayDateTime());
    const [error, setError] = useState('');

    // Get the logged-in user's ID from context or another state management solution
    const { username, isLoggedIn } = useAuth();

    // User should already have been logged in to be able to access Create()
    // const isLoggedIn = username !== null;

    // history for redirect after market creation
    const history = useHistory();

    // Get the timezone offset from the user
    const utcOffset = new Date().getTimezoneOffset();

    const handleSubmit = async (event) => {
        event.preventDefault();
        setError('');

    let isoDateTime = resolutionDateTime; // Default to the original value

    // Convert the resolutionDateTime to ISO format if it's a valid date-time string
    if (resolutionDateTime) {
    const dateTime = new Date(resolutionDateTime);
    if (!isNaN(dateTime.getTime())) {
        // Check if the date object is valid
        isoDateTime = dateTime.toISOString();
    } else {
        console.error('Invalid date-time value:', resolutionDateTime);
        setError('Invalid date-time value');
        return; // Optionally, return early if the date-time is invalid
    }
    }

    // Retrieve the JWT token from localStorage
    const token = localStorage.getItem('token');
    if (!token) {
    alert('Authentication token not found. Please log in again.');
    throw new Error('Token not found');
    }

    try {
    const marketData = {
        questionTitle,
        description,
        outcomeType: 'BINARY',
        resolutionDateTime: isoDateTime,
        initialProbability: 0.5,
        creatorUsername: username,
        isResolved: false,
        utcOffset: utcOffset,
    };

    console.log('marketData:', marketData);

    console.log(JSON.stringify(marketData));

    const response = await fetch(`${API_URL}/api/v0/create`, {
        method: 'POST',
        headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(marketData),
    });

    if (response.ok) {
        const responseData = await response.json();
        console.log('Market creation successful:', responseData);
        const marketId = responseData.id;
        history.push(`/markets/${marketId}`);
    } else {
        const errorText = await response.text(); // Read response as text instead of JSON
        console.error('Market creation failed:', errorText);
        setError('Market creation failed: ' + errorText);
    }
    } catch (error) {
    console.error('Error during market creation:', error);
    setError('Error during market creation: ' + error.message);
    }
};

return (
    <div className='App'>
    <div className='Center-content-header'>
        <h1>Create a Market</h1>
    </div>

    <form onSubmit={handleSubmit} className='create-form'>
        {/* Use RegularInput for Question Title */}
        <div>
        <label>
            Question Title:
            <RegularInput
            type='text'
            value={questionTitle}
            onChange={(e) => setQuestionTitle(e.target.value)}
            />
        </label>
        </div>
        {/* Use RegularInput for Description, ensuring it renders as a textarea */}
        <div>
        <label>
            Description:
            <RegularInputBox
            as='textarea'
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            />
        </label>
        </div>
        {/* Use DatetimeSelector for Resolution Date Time */}
        <div>
        <label>
            Resolution Date Time:
            <DatetimeSelector
            value={resolutionDateTime}

            onChange={(e) => {
                console.log("New date-time:", e.target.value); // Log the new date-time value
                setResolutionDateTime(e.target.value);
            }}

            />
        </label>
        </div>
        {error && <div className='error-message'>{error}</div>}
        <SiteButton type="submit">
                    Create Market
        </SiteButton>
    </form>
    </div>
);
}

export default Create;