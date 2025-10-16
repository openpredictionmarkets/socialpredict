import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import { getEndofDayDateTime } from '../../components/utils/dateTimeTools/FormDateTimeTools';
import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';
import { RegularInput } from '../../components/inputs/InputBar';
import RegularInputBox from '../../components/inputs/InputBox';
import SiteButton from '../../components/buttons/SiteButtons';
import { API_URL } from '../../config';

function Create() {
  const [questionTitle, setQuestionTitle] = useState('');
  const [description, setDescription] = useState('');
  const [resolutionDateTime, setResolutionDateTime] = useState(
    getEndofDayDateTime()
  );
  const [error, setError] = useState('');
  const { username } = useAuth();
  const history = useHistory();

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

    let isoDateTime = resolutionDateTime;

    if (resolutionDateTime) {
      const dateTime = new Date(resolutionDateTime);
      if (!isNaN(dateTime.getTime())) {
        isoDateTime = dateTime.toISOString();
      } else {
        console.error('Invalid date-time value:', resolutionDateTime);
        setError('Invalid date-time value');
        return;
      }
    }

    const token = localStorage.getItem('token');
    if (!token) {
      setError('Authentication token not found. Please log in again.');
      return;
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
        utcOffset: new Date().getTimezoneOffset(),
      };

      console.log('marketData:', marketData);
      console.log(JSON.stringify(marketData));

      const response = await fetch(`${API_URL}/v0/create`, {
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
        history.push(`/markets/${responseData.id}`);
      } else {
        const errorText = await response.text();
        console.error('Market creation failed:', errorText);
        setError(`Market creation failed: ${errorText}`);
      }
    } catch (error) {
      console.error('Error during market creation:', error);
      setError(`Error during market creation: ${error.message}`);
    }
  };

  return (
    <div className='w-full max-w-2xl mx-auto p-4 sm:p-6 bg-gray-800 shadow-lg rounded-lg'>
      <h1 className='text-xl sm:text-2xl font-bold text-white mb-4 sm:mb-6'>
        Create a Market
      </h1>

      <form onSubmit={handleSubmit} className='space-y-4 sm:space-y-6'>
        <div>
          <label className='block text-sm font-medium text-gray-300 mb-1'>
            Question Title
          </label>
          <RegularInput
            type='text'
            value={questionTitle}
            onChange={(e) => setQuestionTitle(e.target.value)}
            placeholder='Enter the market question'
            className='w-full'
          />
        </div>

        <div>
          <label className='block text-sm font-medium text-gray-300 mb-1'>
            Description
          </label>
          <RegularInputBox
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder='Provide details about the market'
            className='h-32 resize-y'
          />
        </div>

        <div>
          <label className='block text-sm font-medium text-gray-300 mb-1'>
            Resolution Date Time
          </label>
          <DatetimeSelector
            value={resolutionDateTime}
            onChange={(e) => {
              console.log('New date-time:', e.target.value);
              setResolutionDateTime(e.target.value);
            }}
            className='w-full'
          />
        </div>

        {error && (
          <div className='bg-red-600 text-white p-3 rounded-md text-sm'>
            {error}
          </div>
        )}

        <SiteButton type='submit' className='w-full'>
          Create Market
        </SiteButton>
      </form>
    </div>
  );
}

export default Create;
