import {
    AssetGroupTagNode,
    AssetGroupTagSelector,
    SeedTypeObjectId,
    SeedTypes,
    SelectorSeedRequest,
} from 'js-client-library';
import { createContext } from 'react';
import { UseQueryResult } from 'react-query';

interface SelectorFormContext {
    seeds: SelectorSeedRequest[];
    setSeeds: React.Dispatch<React.SetStateAction<SelectorSeedRequest[]>>;
    results: AssetGroupTagNode[] | null;
    setResults: React.Dispatch<React.SetStateAction<AssetGroupTagNode[] | null>>;
    selectorType: SeedTypes;
    setSelectorType: React.Dispatch<React.SetStateAction<SeedTypes>>;
    selectorQuery: UseQueryResult<AssetGroupTagSelector>;
}

export const initialValue: SelectorFormContext = {
    seeds: [],
    setSeeds: () => {},
    results: null,
    setResults: () => {},
    selectorType: SeedTypeObjectId,
    setSelectorType: () => {},
    selectorQuery: {
        data: undefined,
        isLoading: true,
        isError: false,
        isSuccess: false,
    } as UseQueryResult<AssetGroupTagSelector>,
};

const SelectorFormContext = createContext<SelectorFormContext>(initialValue);

export default SelectorFormContext;
