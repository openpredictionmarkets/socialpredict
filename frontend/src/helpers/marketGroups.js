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
    const group = market.marketGroup;
    if (!group?.id) {
      groupedRows.push(row);
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
        },
        numUsers: toNumber(row.numUsers),
        totalVolume: toNumber(row.totalVolume),
        marketDust: toNumber(row.marketDust),
      };
      groupsById.set(group.id, aggregate);
      groupedRows.push(aggregate);
      return;
    }

    existing.market.groupChildMarketIds = [
      ...(existing.market.groupChildMarketIds || []),
      market.id || market.marketId,
    ].filter(Boolean);
    existing.market.tags = uniqueTagsBySlug([...(existing.market.tags || []), ...(market.tags || [])]);
    existing.numUsers = Math.max(toNumber(existing.numUsers), toNumber(row.numUsers));
    existing.totalVolume = toNumber(existing.totalVolume) + toNumber(row.totalVolume);
    existing.marketDust = toNumber(existing.marketDust) + toNumber(row.marketDust);
  });

  return groupedRows;
};
