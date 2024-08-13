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
  const [formData, setFormData] = useState({
    questionTitle: '',
    description: '',
    resolutionDateTime: getEndofDayDateTime(),
  });
  const [error, setError] = useState('');
  const { username } = useAuth();
  const history = useHistory();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({ ...prevData, [name]: value }));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');

    const { questionTitle, description, resolutionDateTime } = formData;

    if (!questionTitle.trim() || !description.trim()) {
      setError('Please fill in all fields.');
      return;
    }

    const isoDateTime = new Date(resolutionDateTime).toISOString();
    const utcOffset = new Date().getTimezoneOffset();
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
        utcOffset: utcOffset,
      };

      const response = await fetch(`${API_URL}/api/v0/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(marketData),
      });

      if (response.ok) {
        const { id } = await response.json();
        history.push(`/markets/${id}`);
      } else {
        const errorText = await response.text();
        setError(`Market creation failed: ${errorText}`);
      }
    } catch (error) {
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
            name='questionTitle'
            value={formData.questionTitle}
            onChange={handleChange}
            placeholder='Enter the market question'
            className='w-full'
          />
        </div>

        <div>
          <label className='block text-sm font-medium text-gray-300 mb-1'>
            Description
          </label>
          <RegularInputBox
            name='description'
            value={formData.description}
            onChange={handleChange}
            placeholder='Provide details about the market'
            className='h-32 resize-y'
          />
        </div>

        <div>
          <label className='block text-sm font-medium text-gray-300 mb-1'>
            Resolution Date Time
          </label>
          <DatetimeSelector
            name='resolutionDateTime'
            value={formData.resolutionDateTime}
            onChange={handleChange}
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
