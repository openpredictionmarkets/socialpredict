import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import { getEndofDayDateTime } from '../../components/utils/dateTimeTools/FormDateTimeTools';
import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';
import { RegularInput } from '../../components/inputs/InputBar';
import RegularInputBox from '../../components/inputs/InputBox';
import EmojiPickerInput from '../../components/inputs/EmojiPicker';
import SiteButton from '../../components/buttons/SiteButtons';
import { USER_CREDIT_REFRESH_EVENT } from '../../components/utils/userFinanceTools/FetchUserCredit';
import { apiRequest, authenticatedApiRequest } from '../../api/httpClient';
import { listMarketTags } from '../../api/marketTagsApi';
import MarketTagChips from '../../components/markets/MarketTagChips';

function Create() {
  const [questionTitle, setQuestionTitle] = useState('');
  const [description, setDescription] = useState('');
  const [resolutionDateTime, setResolutionDateTime] = useState(
    getEndofDayDateTime()
  );
  const [yesLabel, setYesLabel] = useState('');
  const [noLabel, setNoLabel] = useState('');
  const [error, setError] = useState('');
  const [createdMarket, setCreatedMarket] = useState(null);
  const [marketCreationCost, setMarketCreationCost] = useState(null);
  const [marketTags, setMarketTags] = useState([]);
  const [selectedTagSlugs, setSelectedTagSlugs] = useState([]);
  const { username } = useAuth();
  const history = useHistory();

  const createMarketReasonMessages = {
    USER_NOT_APPROVED: 'User does not have approval to create markets in moderator mode.',
    AUTHORIZATION_DENIED: 'You are not allowed to create this market.',
    INSUFFICIENT_BALANCE: 'You do not have enough credit to create this market.',
    VALIDATION_FAILED: 'Check the market fields and try again.',
    INVALID_REQUEST: 'Check the market fields and try again.',
  };

  useEffect(() => {
    let ignore = false;

    const loadSetup = async () => {
      try {
        const setup = await apiRequest('/v0/setup');
        const cost = setup?.marketincentives?.createMarketCost;
        if (!ignore && cost !== undefined && cost !== null) {
          setMarketCreationCost(cost);
        }
      } catch {
        // The backend still enforces cost; this call only improves the UI.
      }
    };

    loadSetup();
    return () => {
      ignore = true;
    };
  }, []);

  useEffect(() => {
    let ignore = false;

    const loadTags = async () => {
      try {
        const data = await listMarketTags();
        if (!ignore) {
          setMarketTags(data.tags || []);
        }
      } catch {
        if (!ignore) {
          setMarketTags([]);
        }
      }
    };

    loadTags();
    return () => {
      ignore = true;
    };
  }, []);

  const toggleTagSlug = (slug) => {
    setSelectedTagSlugs((current) => {
      if (current.includes(slug)) {
        return current.filter((value) => value !== slug);
      }
      if (current.length >= 5) {
        setError('You can select up to five market tags.');
        return current;
      }
      setError('');
      return [...current, slug];
    });
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setCreatedMarket(null);

    // Validate custom labels
    const trimmedYesLabel = yesLabel.trim();
    const trimmedNoLabel = noLabel.trim();
    
    if (trimmedYesLabel && (trimmedYesLabel.length < 1 || trimmedYesLabel.length > 20)) {
      setError('Yes label must be between 1 and 20 characters');
      return;
    }
    
    if (trimmedNoLabel && (trimmedNoLabel.length < 1 || trimmedNoLabel.length > 20)) {
      setError('No label must be between 1 and 20 characters');
      return;
    }

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
        yesLabel: trimmedYesLabel || 'YES',
        noLabel: trimmedNoLabel || 'NO',
        tagSlugs: selectedTagSlugs,
      };

      const responseData = await authenticatedApiRequest('/v0/markets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(marketData),
        reasonMessages: createMarketReasonMessages,
        fallbackMessage: 'Market creation failed. Please try again.',
      });

      window.dispatchEvent(new Event(USER_CREDIT_REFRESH_EVENT));
      if (String(responseData.status || '').toLowerCase() === 'proposed') {
        const proposalCost = responseData.proposalCost ?? marketCreationCost;
        setCreatedMarket(responseData);
        history.push('/profile?tab=Proposed%20Markets', {
          proposedMarket: responseData,
          marketCreationCost: proposalCost,
        });
        return;
      }
      history.push(`/markets/${responseData.id}`);
    } catch (error) {
      console.error('Error during market creation:', error);
      setError(error.message || 'Market creation failed. Please try again.');
    }
  };

  return (
    <div className='w-full max-w-2xl mx-auto p-4 sm:p-6 bg-gray-800 shadow-lg rounded-lg'>
      <h1 className='text-xl sm:text-2xl font-bold text-white mb-4 sm:mb-6'>
        Create a Market
      </h1>

      <div className='mb-5 rounded-lg border border-amber-500 bg-amber-950/40 p-4 text-amber-50'>
        <p className='text-sm uppercase tracking-[0.18em] text-amber-300'>
          Market proposal cost
        </p>
        <p className='mt-2 text-2xl font-bold'>
          {marketCreationCost === null ? 'Loading...' : `${marketCreationCost} credits`}
        </p>
        <p className='mt-2 text-sm text-amber-100'>
          This amount is deducted when you create the proposal. If an admin rejects the proposal, the proposal cost is refunded.
        </p>
      </div>

      <form onSubmit={handleSubmit} className='space-y-4 sm:space-y-6'>
        <div>
          <label htmlFor='market-question-title' className='block text-sm font-medium text-gray-300 mb-1'>
            Question Title
          </label>
          <EmojiPickerInput
            id='market-question-title'
            type='text'
            value={questionTitle}
            onChange={(e) => setQuestionTitle(e.target.value)}
            placeholder='Enter the market question'
            className='w-full bg-gray-700 border border-gray-600 text-white px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500'
          />
        </div>

        <div>
          <label htmlFor='market-description' className='block text-sm font-medium text-gray-300 mb-1'>
            Description
          </label>
          <EmojiPickerInput
            id='market-description'
            type='textarea'
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder='Provide details about the market'
            className='w-full h-32 resize-y bg-gray-700 border border-gray-600 text-white px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500'
          />
        </div>

        {marketTags.length > 0 && (
          <div className='rounded-lg border border-gray-700 bg-gray-900/60 p-4'>
            <div className='mb-3 flex flex-wrap items-center justify-between gap-2'>
              <div>
                <p className='text-sm font-medium text-gray-200'>Market Tags</p>
                <p className='mt-1 text-xs text-gray-400'>
                  Pick up to five categories so admins can review routing and users can find this market.
                </p>
              </div>
              <span className='rounded-full bg-gray-800 px-3 py-1 text-xs text-gray-300'>
                {selectedTagSlugs.length}/5 selected
              </span>
            </div>
            <div className='flex flex-wrap gap-2'>
              {marketTags.map((tag) => {
                const selected = selectedTagSlugs.includes(tag.slug);
                return (
                  <button
                    key={tag.slug}
                    type='button'
                    onClick={() => toggleTagSlug(tag.slug)}
                    className={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${
                      selected
                        ? 'border-primary-pink bg-primary-pink/20 text-white'
                        : 'border-gray-600 bg-gray-800 text-gray-300 hover:border-primary-pink/70 hover:text-white'
                    }`}
                  >
                    {tag.displayName || tag.slug}
                  </button>
                );
              })}
            </div>
            <MarketTagChips
              tags={marketTags.filter((tag) => selectedTagSlugs.includes(tag.slug))}
              className='mt-3'
            />
          </div>
        )}

        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          <div>
            <label htmlFor='market-yes-label' className='block text-sm font-medium text-gray-300 mb-1'>
              Yes Label (Optional)
            </label>
            <EmojiPickerInput
              id='market-yes-label'
              type='text'
              value={yesLabel}
              onChange={(e) => setYesLabel(e.target.value)}
              placeholder='e.g., BULL 🚀, WIN, PASS'
              maxLength={20}
              aria-describedby='market-yes-label-hint'
              className='w-full bg-gray-700 border border-gray-600 text-white px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500'
            />
            <p id='market-yes-label-hint' className='text-xs text-gray-400 mt-1'>
              Custom label for positive outcome (defaults to "YES")
            </p>
          </div>
          
          <div>
            <label htmlFor='market-no-label' className='block text-sm font-medium text-gray-300 mb-1'>
              No Label (Optional)
            </label>
            <EmojiPickerInput
              id='market-no-label'
              type='text'
              value={noLabel}
              onChange={(e) => setNoLabel(e.target.value)}
              placeholder='e.g., BEAR 📉, LOSE, FAIL'
              maxLength={20}
              aria-describedby='market-no-label-hint'
              className='w-full bg-gray-700 border border-gray-600 text-white px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500'
            />
            <p id='market-no-label-hint' className='text-xs text-gray-400 mt-1'>
              Custom label for negative outcome (defaults to "NO")
            </p>
          </div>
        </div>

        {(yesLabel.trim() || noLabel.trim()) && (
          <div className='bg-gray-700 p-3 rounded-md' aria-label='Outcome label preview'>
            <p className='text-sm font-medium text-gray-300 mb-2'>Preview:</p>
            <div className='flex space-x-2'>
              <span className='px-3 py-1 bg-green-600 text-white text-sm rounded'>
                {yesLabel.trim() || 'YES'}
              </span>
              <span className='text-gray-400'>vs</span>
              <span className='px-3 py-1 bg-red-600 text-white text-sm rounded'>
                {noLabel.trim() || 'NO'}
              </span>
            </div>
          </div>
        )}

        <div>
          <DatetimeSelector
            id='market-resolution-date-time'
            label='Resolution Date Time'
            value={resolutionDateTime}
            onChange={(e) => setResolutionDateTime(e.target.value)}
            className='w-full'
          />
        </div>

        {error && (
          <div className='bg-red-600 text-white p-3 rounded-md text-sm' role='alert'>
            {error}
          </div>
        )}

        {createdMarket && (
          <div className='rounded-md border border-amber-500 bg-amber-950/40 p-4 text-amber-50'>
            <p className='text-sm uppercase tracking-[0.18em] text-amber-300'>
              Proposed market created
            </p>
            <h2 className='mt-2 text-lg font-semibold'>
              {createdMarket.questionTitle}
            </h2>
            <div className='mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-2'>
              <p>
                <span className='text-amber-200'>Market ID:</span>{' '}
                <span className='font-mono'>{createdMarket.id}</span>
              </p>
              <p>
                <span className='text-amber-200'>Status:</span>{' '}
                <span className='font-mono'>{createdMarket.status}</span>
              </p>
            </div>
            <p className='mt-3 text-sm text-amber-100'>
              This moderator-mode proposal is not tradable until an admin approves it. You will be redirected to your Proposed Markets tab.
            </p>
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
