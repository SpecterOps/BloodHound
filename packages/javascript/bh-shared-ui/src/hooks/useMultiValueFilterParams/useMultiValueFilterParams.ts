import { useSearchParams } from 'react-router-dom';
import { MultiValueFilterConfig, MultiValueSelection, UseMultiValueFilterParams } from './types';
import { areSelectionsEqual, normalizeSelection } from './utils';

/**
 * Synchronizes a multi-value selection with URL search parameters.
 *
 * - Parameter absence resolves to `defaultSelection`
 * - Repeated `valueParam` values represent `some`
 * - `selectionParam=all` represents explicit `all`
 * - `selectionParam=none` represents explicit `none`
 * - Setting the filter to the `defaultSelection` clears both params
 */
export const useMultiValueFilterParams = ({
    valueParam,
    selectionParam,
    defaultSelection,
}: MultiValueFilterConfig): UseMultiValueFilterParams => {
    const [searchParams, setSearchParams] = useSearchParams();

    // Read from URL params into "selection" on render
    const values = searchParams.getAll(valueParam);
    const currentSelectionKind = searchParams.get(selectionParam);

    const normalizedDefaultSelection = normalizeSelection(defaultSelection);

    let selection = normalizedDefaultSelection;

    if (currentSelectionKind === 'all') {
        selection = { kind: 'all' };
    } else if (currentSelectionKind === 'none') {
        selection = { kind: 'none' };
    } else if (values.length > 0) {
        selection = normalizeSelection({ kind: 'some', values });
    }

    const setSelection = (nextSelection: MultiValueSelection) => {
        const normalizedNextSelection = normalizeSelection(nextSelection);

        setSearchParams((currentParams) => {
            const nextParams = new URLSearchParams(currentParams);

            nextParams.delete(valueParam);
            nextParams.delete(selectionParam);

            // Updating to the default value should not recreate either param
            if (areSelectionsEqual(normalizedNextSelection, normalizedDefaultSelection)) {
                return nextParams;
            }

            if (normalizedNextSelection.kind === 'some') {
                normalizedNextSelection.values.forEach((value) => {
                    nextParams.append(valueParam, value);
                });
            } else {
                nextParams.set(selectionParam, normalizedNextSelection.kind);
            }

            return nextParams;
        });
    };

    return { selection, setSelection };
};
