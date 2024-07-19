
export const addOpacityToHex = (hex: string, percent: number): string => {
	if (percent > 100) percent = 100;
	if (percent < 0) percent = 0;

	return `${hex}${Math.floor((percent / 100) * 255)
		.toString(16)
		.padStart(2, '0')}`;
};