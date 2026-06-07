import React, { useState, useEffect } from 'react';
import { SharesBadge, SaleInputAmount, ConfirmSaleButton } from '../../buttons/trade/SellButtons';
import { fetchSaleQuote, fetchUserShares, submitSale } from './TradeUtils';
import { useMarketLabels } from '../../../hooks/useMarketLabels';
import { API_URL } from '../../../config';
import { USER_CREDIT_REFRESH_EVENT } from '../../utils/userFinanceTools/FetchUserCredit';

const SellSharesLayout = ({ marketId, market, token, onTransactionSuccess }) => {
    const [shares, setShares] = useState({ noSharesOwned: 0, yesSharesOwned: 0, value: 0 });
    const [sellAmount, setSellAmount] = useState(1);
    const [selectedOutcome, setSelectedOutcome] = useState(null);
    const [feeData, setFeeData] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [sharesLoading, setSharesLoading] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [saleQuote, setSaleQuote] = useState(null);
    const [quoteError, setQuoteError] = useState('');
    const [isQuoteLoading, setIsQuoteLoading] = useState(false);
    
    // Get custom labels for this market
    const { yesLabel, noLabel } = useMarketLabels(market);
    const showFeeSection = !isLoading && feeData;
    const maxSaleCredits = Math.max(0, Number(shares.value) || 0);

    useEffect(() => {
        const fetchFeeData = async () => {
            if (!token) {
                setIsLoading(false);
                return;
            }

            try {
                const response = await fetch(`${API_URL}/v0/setup`, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });
                if (!response.ok) {
                    throw new Error(`Failed to load setup: ${response.status}`);
                }
                const data = await response.json();
                setFeeData(data.betting?.betFees || null);
            } catch {
                setFeeData(null);
            } finally {
                setIsLoading(false);
            }
        };

        fetchFeeData();
    }, [token]);

    useEffect(() => {
        if (!token) {
            setShares({ noSharesOwned: 0, yesSharesOwned: 0, value: 0 });
            setSelectedOutcome(null);
            setSellAmount(1);
            setSaleQuote(null);
            setQuoteError('');
            return;
        }

        setSharesLoading(true);
        fetchUserShares(marketId, token)
            .then(data => {
                const normalized = normalizeShares(data);
                setShares(normalized);

                // Set outcome and amount based on shares
                if (normalized.noSharesOwned > 0 && normalized.yesSharesOwned === 0) {
                    setSelectedOutcome('NO');
                    setSellAmount(defaultSaleAmount(normalized));
                } else if (normalized.yesSharesOwned > 0 && normalized.noSharesOwned === 0) {
                    setSelectedOutcome('YES');
                    setSellAmount(defaultSaleAmount(normalized));
                } else {
                    setSelectedOutcome(null);
                    setSellAmount(1);
                }
            })
            .catch(error => {
                alert(`Error fetching shares: ${error.message}`);
                // Optionally, reset to default state on error
                setShares({ noSharesOwned: 0, yesSharesOwned: 0, value: 0 });
                setSelectedOutcome(null);
                setSellAmount(1);
            })
            .finally(() => setSharesLoading(false));
    }, [marketId, token]);

    const handleSellAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10) || 0; // Ensure it defaults to 0 if conversion fails
        if (newAmount < 0) {
            return;
        }
        setSaleQuote(null);
        setQuoteError('');
        if (maxSaleCredits > 0 && newAmount > maxSaleCredits) {
            setSellAmount(maxSaleCredits);
            return;
        }
        setSellAmount(newAmount);
    };

    const requestSaleQuote = (outcomeOverride, amountOverride = sellAmount) => {
        const outcomeToUse = outcomeOverride || selectedOutcome;
        if (!outcomeToUse) {
            alert('Please select which shares you would like to sell.');
            return Promise.resolve(null);
        }

        const saleData = {
            marketId,
            outcome: outcomeToUse,
            amount: amountOverride,
        };

        setSelectedOutcome(outcomeToUse);
        setIsQuoteLoading(true);
        setQuoteError('');

        return fetchSaleQuote(saleData, token)
            .then((quote) => {
                setSaleQuote(quote);
                return quote;
            })
            .catch((error) => {
                setSaleQuote(null);
                setQuoteError(error.message);
                return null;
            })
            .finally(() => {
                setIsQuoteLoading(false);
            });
    };

    const handleSaleSubmission = (outcomeOverride) => {
        const outcomeToUse = outcomeOverride || selectedOutcome;
        if (!outcomeToUse) {
            alert('Please select which shares you would like to sell.');
            return;
        }

        const saleData = {
            marketId: marketId,
            outcome: outcomeToUse,
            amount: sellAmount,
        };

        setIsSubmitting(true);
        fetchSaleQuote(saleData, token)
            .then((quote) => {
                setSaleQuote(quote);
                if (!quote.allowed) {
                    alert(quote.message || 'Sale preview is not allowed. Try a different Sale Order amount.');
                    setIsSubmitting(false);
                    return;
                }

                submitSale(
                    saleData,
                    token,
                    (data) => {
                        alert(buildSaleSuccessMessage(data));
                        setSelectedOutcome(null);
                        setSellAmount(1);
                        setIsSubmitting(false);
                        window.dispatchEvent(new Event(USER_CREDIT_REFRESH_EVENT));
                        onTransactionSuccess();
                    },
                    (error) => {
                        alert(`Sale failed: ${error.message}`);
                        setIsSubmitting(false);
                    }
                );
            })
            .catch((error) => {
                alert(`Sale quote failed: ${error.message}`);
                setIsSubmitting(false);
            });
    };

    const isActionDisabled = sharesLoading || isSubmitting || isQuoteLoading;

    return (
        <div className="p-6 bg-blue-900 rounded-lg text-white">
            <h2 className="text-xl mb-4">Shares Owned</h2>
            {shares.noSharesOwned < 1 && shares.yesSharesOwned < 1 ? (
                <div className="text-center">
                    <p>No Shares Owned In This Market</p>
                </div>
            ) : (
                <>
                    <div className="flex justify-center space-x-4 mb-4">
                        {shares.noSharesOwned > 0 &&
                            <SharesBadge type="NO" count={shares.noSharesOwned} label={noLabel} />}
                        {shares.yesSharesOwned > 0 &&
                            <SharesBadge type="YES" count={shares.yesSharesOwned} label={yesLabel} />}
                    </div>
                    {(shares.noSharesOwned > 0 || shares.yesSharesOwned > 0) && (
                        <div className="text-center text-lg mt-2">
                            <span className="font-bold">Position Value: </span>
                            <span className="text-green-300">{shares.value}</span>
                        </div>
                    )}
                    <div className="border-t border-gray-200 my-2"></div>
                    <div className="flex items-center space-x-4 mb-4">
                        <div>
                            <h2 className="text-xl">Sale Order</h2>
                        </div>
                        <SaleInputAmount
                            value={sellAmount}
                            onChange={handleSellAmountChange}
                            max={maxSaleCredits || 1}
                            disabled={isActionDisabled}
                        />
                    </div>
                    <SaleQuotePanel
                        quote={saleQuote}
                        quoteError={quoteError}
                        isLoading={isQuoteLoading}
                        selectedOutcome={selectedOutcome}
                        onSelectAmount={(amount) => {
                            setSellAmount(amount);
                            requestSaleQuote(selectedOutcome, amount);
                        }}
                    />
                    <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
                        {shares.noSharesOwned > 0 &&
                            <SaleActionGroup
                                outcome="NO"
                                disabled={isActionDisabled}
                                isQuoteLoading={isQuoteLoading && selectedOutcome === 'NO'}
                                onTerms={() => requestSaleQuote('NO')}
                                onSubmit={() => {
                                    setSelectedOutcome('NO');
                                    handleSaleSubmission('NO');
                                }}
                            />}
                        {shares.yesSharesOwned > 0 &&
                            <SaleActionGroup
                                outcome="YES"
                                disabled={isActionDisabled}
                                isQuoteLoading={isQuoteLoading && selectedOutcome === 'YES'}
                                onTerms={() => requestSaleQuote('YES')}
                                onSubmit={() => {
                                    setSelectedOutcome('YES');
                                    handleSaleSubmission('YES');
                                }}
                            />}
                    </div>
                </>
            )}

            {showFeeSection && (
                <>
                <div className="border-t border-gray-200 my-2"></div>
                <div className="mb-4">
                    {feeData.initialBetFee === 0 && feeData.sellSharesFee === 0 ? (
                        <p className="text-sm text-gray-300">No fees</p>
                    ) : (
                        <>
                            {feeData.initialBetFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Initial Trade Fee: {feeData.initialBetFee}
                                    <span className="block">Does not apply if already traded on this market.</span>
                                </p>
                            )}
                            {feeData.sellSharesFee > 0 && (
                                <p className="text-sm text-gray-300">
                                    Trading Fee (Selling Share): {feeData.sellSharesFee}
                                </p>
                            )}
                        </>
                    )}
                </div>
                </>
            )}
        </div>
    );


};

const normalizeShares = (data) => {
    if (!data) {
        return { noSharesOwned: 0, yesSharesOwned: 0, value: 0 };
    }
    if (Array.isArray(data)) {
        return normalizeShares(data[0]);
    }

    return {
        noSharesOwned: data.noSharesOwned ?? data.NoSharesOwned ?? 0,
        yesSharesOwned: data.yesSharesOwned ?? data.YesSharesOwned ?? 0,
        value: data.value ?? data.Value ?? 0,
    };
};

const defaultSaleAmount = (normalized) => {
    return Math.max(1, Number(normalized?.value) || 1);
};

const buildSaleSuccessMessage = (data) => {
    const dust = Number(data?.dust) || 0;
    const base = `Sale successful! Sold ${data.sharesSold} shares and credited ${data.saleValue} credits.`;
    if (dust <= 0) {
        return base;
    }
    return `${base} Dust assessed: ${dust} credit${dust === 1 ? '' : 's'} retained by the market due to whole-share rounding.`;
};

const SaleQuotePanel = ({ quote, quoteError, isLoading, selectedOutcome, onSelectAmount }) => {
    if (!selectedOutcome && !quoteError && !isLoading) {
        return null;
    }

    if (isLoading) {
        return (
            <div className="mb-4 rounded-lg border border-blue-700 bg-blue-950/40 p-3 text-sm text-blue-100">
                Calculating sale preview...
            </div>
        );
    }

    if (quoteError) {
        return (
            <div className="mb-4 rounded-lg border border-red-400 bg-red-950/40 p-3 text-sm text-red-100">
                {quoteError}
            </div>
        );
    }

    if (!quote) {
        return null;
    }

    const coverageLabel = Math.round((Number(quote.dustCapCoverage) || 0) * 100);
    const panelTone = quote.allowed
        ? 'border-emerald-400/50 bg-emerald-950/30 text-emerald-50'
        : 'border-amber-300/70 bg-amber-950/40 text-amber-50';

    return (
        <div className={`mb-4 rounded-lg border p-3 text-sm ${panelTone}`}>
            <div className="mb-2 flex items-center justify-between gap-3">
                <h3 className="text-base font-semibold">Sale Preview</h3>
                <span className="rounded-full bg-white/10 px-2 py-1 text-xs">
                    {quote.allowed ? 'Allowed' : 'Adjust amount'}
                </span>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
                <QuoteMetric label="Sale order" value={quote.requestedCredits} />
                <QuoteMetric label="Credits received" value={quote.saleValue} />
                <QuoteMetric label="Shares sold" value={quote.sharesSold} />
                <QuoteMetric label="Dust fee" value={`${quote.dust} / ${quote.maxDust}`} />
                <QuoteMetric label="Value per share" value={quote.valuePerShare} />
                <QuoteMetric label="Dust coverage" value={`${coverageLabel}%`} />
            </div>
            <p className="mt-3 text-xs leading-relaxed">{quote.message}</p>
            <p className="mt-2 text-xs leading-relaxed opacity-80">
                Informational only. Dust fee amounts not guaranteed if trade volume is currently high.
            </p>
            {!quote.allowed && quote.suggestedAmounts?.length > 0 && (
                <div className="mt-3">
                    <p className="mb-2 text-xs font-semibold uppercase tracking-wide">Try a valid amount</p>
                    <div className="flex flex-wrap gap-2">
                        {quote.suggestedAmounts.map((amount) => (
                            <button
                                type="button"
                                key={amount}
                                onClick={() => onSelectAmount(amount)}
                                className="rounded-full bg-white/15 px-3 py-1 text-xs font-semibold hover:bg-white/25"
                            >
                                {amount}
                            </button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

const SaleActionGroup = ({ outcome, disabled, isQuoteLoading, onTerms, onSubmit }) => {
    return (
        <div className="flex w-full flex-col gap-2 rounded-lg border border-white/10 bg-white/5 p-2">
            <ConfirmSaleButton
                onClick={onSubmit}
                selectedDirection={outcome}
                disabled={disabled || isQuoteLoading}
            />
            <button
                type="button"
                onClick={onTerms}
                disabled={disabled || isQuoteLoading}
                className={`w-full rounded border border-blue-200/60 px-4 py-2 text-sm font-semibold text-blue-50 hover:bg-white/10 ${disabled || isQuoteLoading ? 'cursor-not-allowed opacity-50' : ''}`}
            >
                {isQuoteLoading ? 'Loading Terms' : 'Terms'}
            </button>
        </div>
    );
};

const QuoteMetric = ({ label, value }) => (
    <div className="rounded-md bg-white/10 px-3 py-2">
        <p className="text-[0.7rem] uppercase tracking-wide opacity-75">{label}</p>
        <p className="font-semibold">{value}</p>
    </div>
);

export default SellSharesLayout;
