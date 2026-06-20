const toNumber = (value, fallback = 0) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const uniqueTagsBySlug = (tags = []) => {
  const seen = new Set();
  return tags.filter((tag) => {
    const key = tag?.slug || tag?.id || tag?.displayName;
    if (!key || seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
};

const normalizeResolution = (value) => String(value || '').trim().toUpperCase();

const childResolutionFromMarket = (market = {}) => ({
  marketId: market.id || market.marketId,
  answerLabel: market.marketGroup?.answerLabel || market.questionTitle || `Market #${market.id || market.marketId}`,
  isResolved: Boolean(market.isResolved),
  resolutionResult: normalizeResolution(market.resolutionResult),
});

export const summarizeGroupedResolutions = (items = []) => {
  const children = items.filter(Boolean);
  const isResolved = children.length > 0 && children.every((item) => Boolean(item.isResolved));
  if (!isResolved) {
    return null;
  }

  const yesAnswers = children
    .filter((item) => normalizeResolution(item.resolutionResult) === 'YES')
    .map((item) => item.answerLabel)
    .filter(Boolean);

  if (!yesAnswers.length) {
    return {
      isResolved: true,
      yesAnswers,
      label: 'Resolved: no YES answers',
      className: 'text-gray-300',
    };
  }

  return {
    isResolved: true,
    yesAnswers,
    label: `YES: ${yesAnswers.join(', ')}`,
    className: 'text-green-400',
  };
};

export const groupedMarketResolutionSummary = (marketData) => (
  summarizeGroupedResolutions(marketData?.market?.groupChildResolutions || [])
);

export const isGroupedMarketAggregate = (marketData) => (
  Boolean(marketData?.market?.isMarketGroupAggregate && marketData?.market?.marketGroup?.id)
);

export const marketDisplayRoute = (marketData) => {
  const market = marketData?.market || {};
  return market.id || market.marketId ? `/markets/${market.id || market.marketId}` : '#';
};

export const marketProbabilityDisplay = (marketData) => {
  if (isGroupedMarketAggregate(marketData)) {
    return 'Group';
  }
  const probability = toNumber(marketData?.lastProbability, Number.NaN);
  return Number.isFinite(probability) ? probability.toFixed(2) : '—';
};

export const groupedMarketBadgeLabel = (marketData) => {
  const count = toNumber(marketData?.market?.marketGroup?.answerCount, 0);
  if (count <= 0) {
    return 'Grouped market';
  }
  return `${count} answers`;
};

export const groupMarketRows = (marketRows = []) => {
  const groupedRows = [];
  const groupsById = new Map();

  marketRows.forEach((row) => {
    const market = row?.market || {};
    const rowNumUsers = toNumber(row?.numUsers ?? market.numUsers);
    const rowTotalVolume = toNumber(row?.totalVolume ?? market.totalVolume);
    const rowMarketDust = toNumber(row?.marketDust ?? market.marketDust);
    const childResolution = childResolutionFromMarket(market);
    const group = market.marketGroup;
    if (!group?.id) {
      groupedRows.push({
        ...row,
        numUsers: rowNumUsers,
        totalVolume: rowTotalVolume,
        marketDust: rowMarketDust,
      });
      return;
    }

    const existing = groupsById.get(group.id);
    if (!existing) {
      const aggregate = {
        ...row,
        market: {
          ...market,
          id: market.id,
          questionTitle: group.questionTitle || market.questionTitle,
          marketGroup: group,
          isMarketGroupAggregate: true,
          groupChildMarketIds: [market.id || market.marketId].filter(Boolean),
          groupChildResolutions: [childResolution],
        },
        numUsers: rowNumUsers,
        totalVolume: rowTotalVolume,
        marketDust: rowMarketDust,
      };
      groupsById.set(group.id, aggregate);
      groupedRows.push(aggregate);
      return;
    }

    existing.market.groupChildMarketIds = [
      ...(existing.market.groupChildMarketIds || []),
      market.id || market.marketId,
    ].filter(Boolean);
    existing.market.groupChildResolutions = [
      ...(existing.market.groupChildResolutions || []),
      childResolution,
    ];
    existing.market.tags = uniqueTagsBySlug([...(existing.market.tags || []), ...(market.tags || [])]);
    existing.market.isResolved = Boolean(summarizeGroupedResolutions(existing.market.groupChildResolutions));
    existing.numUsers = Math.max(toNumber(existing.numUsers), rowNumUsers);
    existing.totalVolume = toNumber(existing.totalVolume) + rowTotalVolume;
    existing.marketDust = toNumber(existing.marketDust) + rowMarketDust;
  });

  groupedRows.forEach((row) => {
    if (row?.market?.isMarketGroupAggregate) {
      row.market.isResolved = Boolean(summarizeGroupedResolutions(row.market.groupChildResolutions || []));
    }
  });

  return groupedRows;
};
