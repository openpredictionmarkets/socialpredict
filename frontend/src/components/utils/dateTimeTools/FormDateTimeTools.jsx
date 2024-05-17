export function getEndofDayDateTime() {
    const now = new Date();
    now.setHours(23, 59); // Set time to 11:59 PM
    // Format for datetime-local input, which requires 'YYYY-MM-DDTHH:MM'
    return now.toISOString().slice(0, 16);
};

// format the time for the main display grid, showing resolution time
export function formatDateTimeForGrid(dateTimeString) {
    const date = new Date(dateTimeString);

    // Check if there are any seconds
    if (date.getSeconds() > 0) {
      // Add one minute
        date.setMinutes(date.getMinutes() + 1);
      // Reset seconds to zero
        date.setSeconds(0);
    }

    // Extracting date components
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0'); // Months are 0-based
    const day = date.getDate().toString().padStart(2, '0');

    // Convert 24-hour time to 12-hour time and determine AM/PM
    let hour = date.getHours();
    const amPm = hour >= 12 ? 'PM' : 'AM';
    hour = hour % 12;
    hour = hour ? hour : 12; // the hour '0' should be '12'
    const formattedHour = hour.toString().padStart(2, '0');

    const minute = date.getMinutes().toString().padStart(2, '0');

    // Format to YYYY.MM.DD and append time with AM/PM
    return `${year}.${month}.${day} ${formattedHour}:${minute} ${amPm}`;
}