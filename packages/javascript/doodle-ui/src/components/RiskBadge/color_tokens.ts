import { common, palette } from '../../tailwind/colors';

// TODO double check these - there's no dark mode
export const risk = {
    critical: { light: palette.purple.A300 },
    high: { light: palette.red.A300 },
    moderate: { light: palette.brown[300] },
    low: { light: palette.yellow.A300 },
    mitigated: { light: palette.green.A300 },
    resolved: { light: palette['light-blue'].A300 },
    accepted: { light: palette.neutral.light[300] },
    text: { light: common.black },
};
