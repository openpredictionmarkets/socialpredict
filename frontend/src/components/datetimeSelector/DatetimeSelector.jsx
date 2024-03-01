import React, { useState, useEffect } from 'react';

const DatetimeSelector = () => {
    const [selectedDate, setSelectedDate] = useState('');

    useEffect(() => {
        // Construct default date-time string with time set to 11:59 PM
        const now = new Date();
        const year = now.getFullYear();
        const month = now.getMonth() + 1; // getMonth() returns 0-11
        const day = now.getDate();
        // Format month and day to ensure they are in 'MM' and 'DD' format
        const formattedMonth = month < 10 ? `0${month}` : month;
        const formattedDay = day < 10 ? `0${day}` : day;

        // Set default date-time value
        const defaultDateTime = `${year}-${formattedMonth}-${formattedDay}T23:59`;
        setSelectedDate(defaultDateTime);
    }, []);

    const handleChange = (event) => {
        setSelectedDate(event.target.value);
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
                value={selectedDate}
                onChange={handleChange}
            />
        </div>
    );
};

export default DatetimeSelector;
