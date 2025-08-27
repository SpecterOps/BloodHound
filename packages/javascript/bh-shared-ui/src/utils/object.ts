/** Returns Object.entries with results retaining their types */
export const typedEntries = <T extends object>(obj: T): [keyof T, T[keyof T]][] => Object.entries(obj) as any;
