import React from 'react';

const RegularInputBox = ({
  value,
  onChange,
  name,
  placeholder,
  className,
  ...props
}) => {
  return (
    <textarea
      name={name}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      className={`w-full px-4 py-2 border-2 border-blue-500 rounded-md text-white bg-transparent focus:outline-none ${className}`}
      {...props}
    />
  );
};

export default RegularInputBox;
