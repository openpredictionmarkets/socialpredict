import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom/cjs/react-router-dom';
import { API_URL } from '../config';

const DEFAULT_CREATOR_EMOJI = '👤';

const toNumber = (value, fallback = 0) => {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : fallback;
  }
  if (typeof value === 'string') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }
  return fallback;
};

const normalizeProbabilityChange = (change) => {
  if (!change || typeof change !== 'object') {
    return null;
  }

  return {
    probability: toNumber(change.probability ?? change.Probability),
    createdAt: change.createdAt ?? change.CreatedAt ?? null,
    updatedAt: change.updatedAt ?? change.UpdatedAt ?? null,
    txId: change.txId ?? change.TxId,
  };
};

const normalizeProbabilityChanges = (raw) => {
  if (!Array.isArray(raw)) {
    return [];
  }

  return raw
    .map(normalizeProbabilityChange)
    .filter((item) => item !== null);
};

const normalizeMarket = (market) => {
  if (!market || typeof market !== 'object') {
    return {
      id: null,
      questionTitle: 'Untitled market',
      description: '',
      outcomeType: '',
      resolutionDateTime: null,
      creatorUsername: 'unknown',
      yesLabel: '',
      noLabel: '',
      status: '',
      createdAt: null,
      updatedAt: null,
      initialProbability: 0,
      isResolved: false,
      resolutionResult: null,
    };
  }

  return {
    id: market.id ?? market.ID ?? null,
    questionTitle: market.questionTitle ?? market.QuestionTitle ?? 'Untitled market',
    description: market.description ?? market.Description ?? '',
    outcomeType: market.outcomeType ?? market.OutcomeType ?? '',
    resolutionDateTime: market.resolutionDateTime ?? market.ResolutionDateTime ?? null,
    creatorUsername: market.creatorUsername ?? market.CreatorUsername ?? 'unknown',
    yesLabel: market.yesLabel ?? market.YesLabel ?? '',
    noLabel: market.noLabel ?? market.NoLabel ?? '',
    status: market.status ?? market.Status ?? '',
    createdAt: market.createdAt ?? market.CreatedAt ?? null,
    updatedAt: market.updatedAt ?? market.UpdatedAt ?? null,
    initialProbability: toNumber(market.initialProbability ?? market.InitialProbability),
    isResolved: market.isResolved ?? market.IsResolved ?? false,
    resolutionResult: market.resolutionResult ?? market.ResolutionResult ?? null,
  };
};

const normalizeCreator = (creator, fallbackUsername) => {
  if (!creator || typeof creator !== 'object') {
    return {
      username: fallbackUsername ?? 'unknown',
      personalEmoji: DEFAULT_CREATOR_EMOJI,
    };
  }

  return {
    username: creator.username ?? creator.Username ?? fallbackUsername ?? 'unknown',
    personalEmoji: creator.personalEmoji ?? creator.PersonalEmoji ?? DEFAULT_CREATOR_EMOJI,
    displayName: creator.displayName ?? creator.DisplayName,
  };
};

const normalizeMarketDetails = (raw) => {
  if (!raw || typeof raw !== 'object') {
    return null;
  }

  const normalizedMarket = normalizeMarket(raw.market ?? raw.Market);
  const normalizedCreator = normalizeCreator(raw.creator ?? raw.Creator, normalizedMarket.creatorUsername);

  return {
    market: normalizedMarket,
    creator: normalizedCreator,
    probabilityChanges: normalizeProbabilityChanges(raw.probabilityChanges ?? raw.ProbabilityChanges),
    numUsers: toNumber(raw.numUsers ?? raw.NumUsers),
    totalVolume: toNumber(raw.totalVolume ?? raw.TotalVolume),
    marketDust: toNumber(raw.marketDust ?? raw.MarketDust),
    lastProbability: toNumber(raw.lastProbability ?? raw.LastProbability),
  };
};

const calculateCurrentProbability = (details) => {
  if (!details) return 0;

  const changes = Array.isArray(details.probabilityChanges)
    ? details.probabilityChanges
    : [];

  if (changes.length > 0) {
    const last = changes[changes.length - 1];
    const probability = toNumber(last.probability, details.lastProbability);
    return parseFloat(probability.toFixed(3));
  }

  const baseProbability = toNumber(
    details.lastProbability ?? details.market?.initialProbability,
  );

  return parseFloat(baseProbability.toFixed(3));
};

export const useMarketDetails = () => {
  const [details, setDetails] = useState(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState(null);
  const [currentProbability, setCurrentProbability] = useState(0);
  const { marketId } = useParams();
  const [triggerRefresh, setTriggerRefresh] = useState(false);

  useEffect(() => {
    const fetchedToken = localStorage.getItem('token');
    setToken(fetchedToken);
    setIsLoggedIn(!!fetchedToken);
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(`${API_URL}/v0/markets/${marketId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch market data');
        }
        const data = await response.json();
        const normalized = normalizeMarketDetails(data);
        setDetails(normalized);
        setCurrentProbability(calculateCurrentProbability(normalized));
      } catch (error) {
        console.error('Error fetching market data:', error);
      }
    };

    fetchData();
  }, [marketId, triggerRefresh]);

  const refetchData = () => {
    setTriggerRefresh((prev) => !prev);
  };

  return { details, isLoggedIn, token, refetchData, currentProbability };
};
