export const MIN_SIG_FIGS = 2;
export const MAX_SIG_FIGS = 9;
export const DEFAULT_SIG_FIGS = 4;

export const clampSigFigs = (value) => {
  if (!Number.isFinite(value)) {
    return DEFAULT_SIG_FIGS;
  }

  if (value < MIN_SIG_FIGS) {
    return MIN_SIG_FIGS;
  }

  if (value > MAX_SIG_FIGS) {
    return MAX_SIG_FIGS;
  }

  return Math.round(value);
};

export const createFormatter = (sigFigs) => (rawValue) => {
  if (rawValue === null || rawValue === undefined) {
    return '';
  }

  const numericValue = Number(rawValue);
  if (!Number.isFinite(numericValue)) {
    return '';
  }

  return numericValue.toPrecision(sigFigs);
};

export const loadChartFormatter = async () => {
  let requestedSigFigs = DEFAULT_SIG_FIGS;

  try {
    const response = await fetch('/v0/setup/frontend', {
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      const data = await response.json();
      requestedSigFigs = Number(data?.charts?.sigFigs ?? DEFAULT_SIG_FIGS);
    }
  } catch (error) {
    console.error('Failed to load chart formatter config', error);
  }

  const sigFigs = clampSigFigs(requestedSigFigs);
  return {
    sigFigs,
    format: createFormatter(sigFigs),
  };
};

export const SocialPredictChartTools = {
  clampSigFigs,
  createFormatter,
  loadChartFormatter,
};
