export type MultiValueSelection = { kind: 'all' } | { kind: 'some'; values: string[] } | { kind: 'none' };

export type MultiValueFilterConfig = {
    valueParam: string;
    selectionParam: string;
    defaultSelection: MultiValueSelection;
};

export type UseMultiValueFilterParams = {
    selection: MultiValueSelection;
    setSelection: (selection: MultiValueSelection) => void;
};
