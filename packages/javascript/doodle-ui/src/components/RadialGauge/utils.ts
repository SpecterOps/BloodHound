export const getCircumference = (radius: number) => {
    return 2 * Math.PI * radius;
};

export const clampNumber = (value: number, lowerLimit: number, upperLimit: number) => {
    if (lowerLimit > upperLimit)
        console.warn(`clampNumber limits appear inconsistent: lowerLimit ${lowerLimit} upperLimit ${upperLimit}`);
    return Math.max(lowerLimit, Math.min(value, upperLimit));
};
