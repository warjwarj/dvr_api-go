
export function getCurrentTime() {
    // Get the current date and time
    let currentTime = new Date();

    // Extract individual components of the current time
    let hours = currentTime.getHours();
    let minutes = currentTime.getMinutes();
    let seconds = currentTime.getSeconds();

    // Pad single-digit values with leading zeros
    hours = hours < 10 ? '0' + hours : hours;
    minutes = minutes < 10 ? '0' + minutes : minutes;
    seconds = seconds < 10 ? '0' + seconds : seconds;

    // Format the current time
    let formattedTime = `${hours}:${minutes}:${seconds}`;

    // Return the formatted time
    return formattedTime;
}

// format frmo html format into the format the commadns take
export function formatDateTimeDVRFormat(inputDateTime) {
    return inputDateTime.split('T').map(part => part.replace(/[-:]/g, '')).join('-') + "00";
}