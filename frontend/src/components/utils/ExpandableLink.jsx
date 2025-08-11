import React, { useState } from 'react';
import { Link } from 'react-router-dom';

const ExpandableLink = ({
  text,
  maxLength = 50,
  to,
  className = '',
  buttonClassName = '',
  expandIcon = '📐',
  linkClassName = ''
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  if (!text || text.length <= maxLength) {
    return (
      <Link to={to} className={linkClassName}>
        {text}
      </Link>
    );
  }

  const truncatedText = text.slice(0, maxLength) + '...';

  return (
    <div className={className}>
      <Link to={to} className={linkClassName}>
        {isExpanded ? text : truncatedText}
      </Link>
      <button
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setIsExpanded(!isExpanded);
        }}
        className={`inline-block ${buttonClassName}`}
        title={isExpanded ? 'Collapse' : 'Expand'}
        type="button"
      >
        {expandIcon} {isExpanded ? 'Less' : 'More'}
      </button>
    </div>
  );
};

export default ExpandableLink;
