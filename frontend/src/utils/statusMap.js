/**
 * Mapping from UI tab labels to backend status values
 * This ensures consistency across all components that need to translate
 * between UI tab names and API status parameters
 */
export const TAB_TO_STATUS = {
  Active: "active",
  Closed: "closed", 
  Resolved: "resolved",
  All: "all", // Backend interprets "all" as no filter
};

/**
 * Reverse mapping from status to tab labels
 * Useful for determining which tab should be active based on status
 */
export const STATUS_TO_TAB = {
  active: "Active",
  closed: "Closed",
  resolved: "Resolved", 
  all: "All",
};
