import React, { useState } from 'react';

const ExpandableText = ({ 
  text, 
  maxLength = 50,
  className = '',
  expandedClassName = 'mt-2 p-2 bg-gray-700 rounded border border-gray-600',
  buttonClassName = 'text-xs text-blue-400 hover:text-blue-300 transition-colors ml-1',
  showFullTextInExpanded = true,
  expandIcon = '📐'
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  // If text is shorter than maxLength, no need for expansion
  if (text.length <= maxLength) {
    return <span className={className}>{text}</span>;
  }

  const truncatedText = text.slice(0, maxLength) + '...';

  return (
    <span className={className}>
      {isExpanded ? text : truncatedText}
      <button
        onClick={(e) => {
          e.preventDefault(); // Prevent navigation if inside a Link
          e.stopPropagation(); // Prevent event bubbling
          setIsExpanded(!isExpanded);
        }}
        className={buttonClassName}
        title={isExpanded ? "Show less" : "Show full text"}
      >
        {expandIcon}
      </button>
      {isExpanded && showFullTextInExpanded && (
        <div className={expandedClassName}>
          <span className="text-gray-300 text-sm">{text}</span>
        </div>
      )}
    </span>
  );
};

export default ExpandableText;
