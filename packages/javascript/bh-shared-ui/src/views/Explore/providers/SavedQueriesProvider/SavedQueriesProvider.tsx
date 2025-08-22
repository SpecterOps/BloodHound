import { createContext, Dispatch, SetStateAction, useContext, useState } from 'react';
import { useCypherSearch, useGetSelectedQuery } from '../../../../hooks';
import { QueryLineItem } from '../../../../types';
type SelectedType = {
    query: string;
    id?: number;
};
type SaveAction = 'edit' | 'save-as' | undefined;

interface SavedQueriesContextType {
    selected: SelectedType;
    selectedQuery: QueryLineItem | undefined;
    showSaveQueryDialog: boolean;
    saveAction: SaveAction;
    setSelected: Dispatch<SetStateAction<SelectedType>>;
    setShowSaveQueryDialog: Dispatch<SetStateAction<boolean>>;
    setSaveAction: Dispatch<SetStateAction<SaveAction>>;
    runQuery: any;
    editQuery: any;
}

export const SavedQueriesContext = createContext<SavedQueriesContextType>({
    selected: { query: '', id: undefined },
    selectedQuery: undefined,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: (value) => {},
    setShowSaveQueryDialog: () => {},
    runQuery: (query: string, id: number) => {},
    editQuery: (id: number) => {},
    setSaveAction: (value) => {},
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
    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);
    const [saveAction, setSaveAction] = useState<SaveAction>(undefined);

    const selectedQuery: QueryLineItem | undefined = useGetSelectedQuery(selected.query, selected.id);

    const { setCypherQuery, performSearch } = useCypherSearch();

    const runQuery = (query: string, id: number) => {
        setSelected({ query, id });
        setCypherQuery(query);
        performSearch(query);
    };

    const editQuery = (id: number) => {
        setSelected({ query: '', id: id });
        setSaveAction('edit');
        setShowSaveQueryDialog(true);
    };

    const contextValue = {
        selected,
        selectedQuery,
        saveAction,
        showSaveQueryDialog,
        setSelected,
        setSaveAction,
        setShowSaveQueryDialog,
        runQuery,
        editQuery,
    };

    return <SavedQueriesContext.Provider value={contextValue}>{children}</SavedQueriesContext.Provider>;
};
