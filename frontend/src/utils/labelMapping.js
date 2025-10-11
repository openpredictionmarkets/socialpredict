/**
 * Label Mapping Utilities
 * 
 * These utilities handle the mapping between internal YES/NO values
 * and custom display labels while preserving business logic.
 */

/**
 * Get the display labels for a market
 * @param {Object} market - Market object with yesLabel and noLabel fields
 * @returns {Object} Object with yes and no display labels
 */
export const getMarketLabels = (market) => {
  if (!market) {
    return { yes: "YES", no: "NO" };
  }
  
  return {
    yes: market.yesLabel || "YES",
    no: market.noLabel || "NO"
  };
};

/**
 * Map internal YES/NO values to display labels
 * @param {string} internalValue - Internal value ("YES", "NO", or other)
 * @param {Object} market - Market object with label fields
 * @returns {string} Display label for the internal value
 */
export const mapInternalToDisplay = (internalValue, market) => {
  if (!internalValue || !market) {
    return internalValue || "";
  }
  
  const upperValue = internalValue.toUpperCase();
  
  if (upperValue === "YES") {
    return market.yesLabel || "YES";
  }
  
  if (upperValue === "NO") {
    return market.noLabel || "NO";
  }
  
  // Return original value if not YES/NO
  return internalValue;
};

/**
 * Map display label back to internal YES/NO value
 * @param {string} displayValue - Display label
 * @param {Object} market - Market object with label fields
 * @returns {string} Internal value ("YES" or "NO")
 */
export const mapDisplayToInternal = (displayValue, market) => {
  if (!displayValue || !market) {
    return displayValue || "";
  }
  
  // Check if display value matches YES label
  if (displayValue === market.yesLabel) {
    return "YES";
  }
  
  // Check if display value matches NO label
  if (displayValue === market.noLabel) {
    return "NO";
  }
  
  // Fallback: if it's already YES/NO, return as-is
  const upperValue = displayValue.toUpperCase();
  if (upperValue === "YES" || upperValue === "NO") {
    return upperValue;
  }
  
  // Return original value if no mapping found
  return displayValue;
};

/**
 * Get the appropriate CSS class for a result based on internal value
 * @param {string} internalValue - Internal value ("YES" or "NO")
 * @returns {string} CSS class name
 */
export const getResultCssClass = (internalValue) => {
  if (!internalValue) return "";
  
  const upperValue = internalValue.toUpperCase();
  
  if (upperValue === "YES") {
    return "text-green-400";
  }
  
  if (upperValue === "NO") {
    return "text-red-400";
  }
  
  return "";
};

/**
 * Get the resolved text with custom labels
 * @param {string} internalValue - Internal resolution result ("YES" or "NO")
 * @param {Object} market - Market object with label fields
 * @returns {string} Resolved text with custom labels
 */
export const getResolvedText = (internalValue, market) => {
  if (!internalValue || !market) {
    return "Pending";
  }
  
  const displayLabel = mapInternalToDisplay(internalValue, market);
  return `Resolved ${displayLabel}`;
};

/**
 * Validate that market has required label fields
 * @param {Object} market - Market object to validate
 * @returns {boolean} True if market has valid label fields
 */
export const hasValidLabels = (market) => {
  return market && 
         typeof market.yesLabel === 'string' && 
         typeof market.noLabel === 'string' &&
         market.yesLabel.trim() !== '' &&
         market.noLabel.trim() !== '';
};

/**
 * Get fallback labels if market labels are invalid
 * @param {Object} market - Market object
 * @returns {Object} Object with fallback labels
 */
export const getFallbackLabels = (market) => {
  const labels = getMarketLabels(market);
  
  // If labels are empty or just whitespace, use defaults
  return {
    yes: labels.yes.trim() || "YES",
    no: labels.no.trim() || "NO"
  };
};