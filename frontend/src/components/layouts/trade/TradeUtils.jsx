import {
  fetchUserPosition,
  NO_SELLABLE_SHARES_MESSAGE,
  placeBet,
  quoteSale,
  sellShares,
} from '../../../api/tradeApi';

export { NO_SELLABLE_SHARES_MESSAGE };

export const submitBet = (betData, token, onSuccess, onError) => {
  placeBet({ token, ...betData })
    .then(onSuccess)
    .catch(onError);
};

export const fetchUserShares = async (marketId, token) => fetchUserPosition({ token, marketId });

export const fetchSaleQuote = async (saleData, token) => quoteSale({ token, ...saleData });

export const submitSale = (saleData, token, onSuccess, onError) => {
  sellShares({ token, ...saleData })
    .then(onSuccess)
    .catch(onError);
};
