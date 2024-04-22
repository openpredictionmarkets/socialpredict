const getEndofDayDateTime = () => {
    const now = new Date();
    now.setHours(23, 59); // Set time to 11:59 PM
    // Format for datetime-local input, which requires 'YYYY-MM-DDTHH:MM'
    return now.toISOString().slice(0, 16);
};

export default getEndofDayDateTime;