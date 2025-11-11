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

const buildFormatterState = (sigFigs) => ({
  sigFigs,
  format: createFormatter(sigFigs),
});

let cachedFormatter = buildFormatterState(DEFAULT_SIG_FIGS);
let formatterPromise = null;

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

      const rawSigFigs = data?.charts?.sigFigs;

      requestedSigFigs = Number.isFinite(Number(rawSigFigs))
        ? Number(rawSigFigs)
        : DEFAULT_SIG_FIGS;
      
    }
  } catch (error) {
    console.error('Failed to load chart formatter config', error);
  }

  const sigFigs = clampSigFigs(requestedSigFigs);
  return buildFormatterState(sigFigs);
};

export const ensureChartFormatter = () => {
  if (!formatterPromise) {
    formatterPromise = loadChartFormatter().then((formatter) => {
      cachedFormatter = formatter;
      return formatter;
    });
  }
  return formatterPromise;
};

export const getChartFormatter = () => cachedFormatter;

export const formatChartValue = (value) => cachedFormatter.format(value);

export const SocialPredictChartFormatter = {
  ensure: ensureChartFormatter,
  format: formatChartValue,
  getFormatter: getChartFormatter,
  getSigFigs: () => cachedFormatter.sigFigs,
};

export const SocialPredictChartTools = {
  clampSigFigs,
  createFormatter,
  loadChartFormatter,
  ensureChartFormatter,
  formatChartValue,
  getChartFormatter,
};
