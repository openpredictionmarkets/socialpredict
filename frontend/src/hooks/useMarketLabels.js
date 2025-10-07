import { useMemo } from 'react';
import { 
  getMarketLabels, 
  mapInternalToDisplay, 
  mapDisplayToInternal,
  getResultCssClass,
  getResolvedText,
  hasValidLabels,
  getFallbackLabels
} from '../utils/labelMapping';

/**
 * React hook for market label mapping
 * Provides easy access to label mapping utilities with memoization
 * @param {Object} market - Market object with label fields
 * @returns {Object} Object with label mapping functions and values
 */
export const useMarketLabels = (market) => {
  const labels = useMemo(() => getFallbackLabels(market), [market]);
  
  const mapToDisplay = useMemo(
    () => (internalValue) => mapInternalToDisplay(internalValue, market),
    [market]
  );
  
  const mapToInternal = useMemo(
    () => (displayValue) => mapDisplayToInternal(displayValue, market),
    [market]
  );
  
  const getResultClass = useMemo(
    () => (internalValue) => getResultCssClass(internalValue),
    []
  );
  
  const getResolvedLabel = useMemo(
    () => (internalValue) => getResolvedText(internalValue, market),
    [market]
  );
  
  const isValidLabels = useMemo(() => hasValidLabels(market), [market]);
  
  return {
    // Labels for display
    yesLabel: labels.yes,
    noLabel: labels.no,
    
    // Mapping functions
    mapToDisplay,
    mapToInternal,
    getResultClass,
    getResolvedLabel,
    
    // Validation
    isValidLabels,
    
    // Original market for reference
    market
  };
};

export default useMarketLabels;