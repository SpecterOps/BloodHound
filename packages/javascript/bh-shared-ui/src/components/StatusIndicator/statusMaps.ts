type StatusMapping = Record<string, { name: string; color: string }>;

const COLOR_BAD = `fill-[#D9442E]`; // Red
const COLOR_GOOD = `fill-[#BCD3A8]`; // Olive Green
const COLOR_PENDING = `fill-[#5CC3AD] animate-pulse`; // Aqua
const COLOR_UNKNOWN = `fill-[#5CC3AD]`; // Aqua

export const DEFAULT_MAP: StatusMapping = {
    0: {
        name: 'Neutral',
        color: COLOR_UNKNOWN,
    },
    1: {
        name: 'Good',
        color: COLOR_GOOD,
    },
    2: {
        name: 'Bad',
        color: COLOR_BAD,
    },
};

export const JOB_STATUS_MAP: StatusMapping = {
    [-1]: {
        color: COLOR_BAD,
        name: 'Invalid',
    },
    0: {
        color: COLOR_GOOD,
        name: 'Ready',
    },
    1: {
        color: COLOR_PENDING,
        name: 'Running',
    },
    2: {
        color: COLOR_GOOD,
        name: 'Complete',
    },
    3: {
        color: COLOR_BAD,
        name: 'Canceled',
    },
    4: {
        color: COLOR_BAD,
        name: 'Timed Out',
    },
    5: {
        color: COLOR_BAD,
        name: 'Failed',
    },
    6: {
        color: COLOR_PENDING,
        name: 'Ingesting',
    },
    7: {
        color: COLOR_PENDING,
        name: 'Analyzing',
    },
    8: {
        color: COLOR_UNKNOWN,
        name: 'Partially Completed',
    },
};
