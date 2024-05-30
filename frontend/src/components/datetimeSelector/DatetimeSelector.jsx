import React, { useEffect, useState } from 'react';

const DatetimeSelector = ({ value, onChange }) => {
    const [internalValue, setInternalValue] = useState(value);

    // Set internalValue when component mounts
    useEffect(() => {
        if (!value) { // Only set default if no value is provided
            const now = new Date();
            const year = now.getFullYear();
            const month = now.getMonth() + 1;
            const day = now.getDate();
            const formattedMonth = month < 10 ? `0${month}` : `${month}`;
            const formattedDay = day < 10 ? `0${day}` : `${day}`;
            const defaultDateTime = `${year}-${formattedMonth}-${formattedDay}T23:59`;
            setInternalValue(defaultDateTime);
        }
    }, [value]);

    const handleChange = (event) => {
        setInternalValue(event.target.value);
        onChange(event); // Propagate changes to parent
    };

    return (
        <div className="p-4 bg-custom-gray-light text-white rounded-lg shadow-md max-w-md mx-auto my-4">
            <label htmlFor="datetime-selector" className="block mb-2 font-bold">
                Select Date and Time:
            </label>
            <input
                id="datetime-selector"
                type="datetime-local"
                className="w-full p-2 rounded border-gray-300 shadow-sm bg-white text-black"
                value={internalValue}
                onChange={handleChange}
            />
        </div>
    );
};

export default DatetimeSelector;