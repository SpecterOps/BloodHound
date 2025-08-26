import { createContext, Dispatch, SetStateAction, useContext } from 'react';
import { QueryLineItem, SaveQueryAction, SelectedQuery } from '../../../../types';

interface SavedQueriesContextType {
    selected: SelectedQuery;
    selectedQuery: QueryLineItem | undefined;
    showSaveQueryDialog: boolean;
    saveAction: SaveQueryAction;
    setSelected: Dispatch<SetStateAction<SelectedQuery>>;
    setShowSaveQueryDialog: Dispatch<SetStateAction<boolean>>;
    setSaveAction: Dispatch<SetStateAction<SaveQueryAction>>;
    runQuery: any;
    editQuery: any;
}

export const SavedQueriesContext = createContext<SavedQueriesContextType>({
    selected: { query: '', id: undefined },
    selectedQuery: undefined,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: () => {},
    setShowSaveQueryDialog: () => {},
    runQuery: () => {},
    editQuery: () => {},
    setSaveAction: () => {},
});

export const useSavedQueriesContext = () => {
    const context = useContext(SavedQueriesContext);
    if (!context) {
        throw new Error('MyContext provider is missing!');
    }
    return context;
};
