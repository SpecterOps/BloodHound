import { createContext, Dispatch, SetStateAction, useContext, useState } from 'react';
import { useCypherSearch, useGetSelectedQuery } from '../../../../hooks';
import { QueryLineItem } from '../../../../types';
type SelectedType = {
    query: string;
    id?: number;
};
interface SavedQueriesContextType {
    selected: SelectedType;
    selectedQuery: QueryLineItem | undefined;
    setSelected: Dispatch<SetStateAction<SelectedType>>;
    runQuery: any;
}

type CypherSearchState = {
    cypherQuery: string;
    setCypherQuery: (query: string) => void;
    performSearch: (query?: string) => void;
};

export const SavedQueriesContext = createContext<SavedQueriesContextType>({
    selected: { query: '', id: undefined },
    selectedQuery: undefined,
    setSelected: (value) => {},
    runQuery: (query: string, id: number) => {},
});

export const useSavedQueriesContext = () => {
    const context = useContext(SavedQueriesContext);
    if (!context) {
        throw new Error('MyContext provider is missing!');
    }
    return context;
};

export const SavedQueriesProvider = ({ children }: { children: any }) => {
    const [selected, setSelected] = useState<SelectedType>({ query: '', id: undefined });
    const selectedQuery: QueryLineItem | undefined = useGetSelectedQuery(selected.query, selected.id);

    const { cypherQuery, setCypherQuery, performSearch } = useCypherSearch();

    const runQuery = (query: string, id: number) => {
        setSelected({ query, id });
        setCypherQuery(query);
        performSearch(query);
    };

    return (
        <SavedQueriesContext.Provider value={{ selected, selectedQuery, setSelected, runQuery }}>
            {children}
        </SavedQueriesContext.Provider>
    );
};
