export const abbreviatedNumber = (num: number, fractionDigits: number = 1) => {
    if (num < 1000) {
        // If the number is less than 1000, no abbreviation needed
        return num.toString();
    }
    const abbreviations = ['', 'K', 'M', 'B', 'T'];
    const log1000 = Math.floor(Math.log10(Math.abs(num)) / 3); // appropriate abbreviation index

    // Otherwise, divide the number by the appropriate power of 1000 and add the abbreviation
    const formattedNumber = (num / Math.pow(1000, log1000)).toFixed(fractionDigits);
    return formattedNumber + abbreviations[log1000];
};
