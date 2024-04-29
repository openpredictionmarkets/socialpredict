const FormatDateForAxis = (unixTime) => {
    const date = new Date(unixTime);
    const day = date.getDate();
    const month = date.toLocaleString('en-US', { month: 'short' }); // 'short' gives the three-letter abbreviation
    return `${day}-${month.toUpperCase()}`; // Formats the date as DD-MMM
};

export default FormatDateForAxis;