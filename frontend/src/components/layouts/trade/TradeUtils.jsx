import { API_URL } from '../../../config';

const unwrapApiResponse = (payload) => {
    if (payload && typeof payload === 'object' && 'ok' in payload) {
        if (payload.ok === false) {
            throw new Error(payload.reason || 'Request failed');
        }

        if (payload.ok === true && 'result' in payload) {
            return payload.result;
        }
    }

    return payload;
};

const formatApiError = (errorData, fallback) => {
    if (!errorData || typeof errorData !== 'object') {
        return fallback;
    }

    if (errorData.reason === 'DUST_CAP_EXCEEDED') {
        const dust = errorData.details?.dust;
        const maxDust = errorData.details?.maxDust;
        const suffix = dust !== undefined && maxDust !== undefined
            ? ` This sale would create ${dust} dust, but the maximum allowed is ${maxDust}.`
            : '';
        return `${errorData.message || 'Sale would create too much rounding dust.'}${suffix} Try a different requested credit amount.`;
    }

    return errorData.message || errorData.reason || errorData.error || fallback;
};

const parseErrorResponse = async (response, fallbackPrefix) => {
    const text = await response.text();
    let errorMessage;
    try {
        const errorData = JSON.parse(text);
        errorMessage = formatApiError(errorData, text);
    } catch {
        errorMessage = text || `HTTP ${response.status}: ${response.statusText}`;
    }
    throw new Error(`${fallbackPrefix} (${response.status}): ${errorMessage}`);
};

export const submitBet = (betData, token, onSuccess, onError) => {
    if (!token) {
        alert('Please log in to place a bet.');
        return;
    }

    if (!betData.marketId || !betData.amount || !betData.outcome) {
        onError(new Error('Missing required bet data (marketId, amount, outcome)'));
        return;
    }

    if (betData.amount < 1) {
        onError(new Error('Bet amount must be at least 1'));
        return;
    }

    if (betData.outcome !== 'YES' && betData.outcome !== 'NO') {
        onError(new Error('Bet outcome must be YES or NO'));
        return;
    }

    fetch(`${API_URL}/v0/bet`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(betData),
    })
    .then(response => {
        if (!response.ok) {
            return parseErrorResponse(response, 'Bet failed');
        }
        return response.json();
    })
    .then(data => onSuccess(unwrapApiResponse(data)))
    .catch(error => {
        onError(error);
    });
};

export const fetchUserShares = async (marketId, token) => {
    const response = await fetch(`${API_URL}/v0/userposition/${marketId}`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    if (!response.ok) {
        throw new Error('Failed to fetch user shares');
    }
    const data = await response.json();
    return unwrapApiResponse(data);
};

export const fetchSaleQuote = async (saleData, token) => {
    if (!token) {
        throw new Error('Please log in to sell shares.');
    }

    if (!saleData.marketId || !saleData.amount || !saleData.outcome) {
        throw new Error('Missing required sale data (marketId, amount, outcome)');
    }

    if (saleData.amount < 1) {
        throw new Error('Sale amount must be at least 1');
    }

    if (saleData.outcome !== 'YES' && saleData.outcome !== 'NO') {
        throw new Error('Sale outcome must be YES or NO');
    }

    const response = await fetch(`${API_URL}/v0/sell/quote`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(saleData),
    });

    if (!response.ok) {
        return parseErrorResponse(response, 'Sale quote failed');
    }

    return unwrapApiResponse(await response.json());
};

export const submitSale = (saleData, token, onSuccess, onError) => {
    if (!token) {
        alert('Please log in to sell shares.');
        return;
    }

    if (!saleData.marketId || !saleData.amount || !saleData.outcome) {
        onError(new Error('Missing required sale data (marketId, amount, outcome)'));
        return;
    }

    if (saleData.amount < 1) {
        onError(new Error('Sale amount must be at least 1'));
        return;
    }

    if (saleData.outcome !== 'YES' && saleData.outcome !== 'NO') {
        onError(new Error('Sale outcome must be YES or NO'));
        return;
    }

    fetch(`${API_URL}/v0/sell`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(saleData),
    })
    .then(response => {
        if (!response.ok) {
            return parseErrorResponse(response, 'Sale failed');
        }
        return response.json();
    })
    .then(data => onSuccess(unwrapApiResponse(data)))
    .catch(error => {
        onError(error);
    });
};
