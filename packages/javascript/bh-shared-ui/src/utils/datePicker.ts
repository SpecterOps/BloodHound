export enum CustomRangeError {
    INVALID_DATE = 'Must be a valid date in yyyy-mm-dd format.',
    INVALID_RANGE_START = 'Start date must be before end date.',
    INVALID_RANGE_END = 'End date must be after start date.',
}

export const START_DATE = 'start-date' as const;
export const END_DATE = 'end-date' as const;
