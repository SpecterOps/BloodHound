import { createContext, useContext } from 'react';

export interface HistoryTableContext {
    currentNote: any;
    setCurrentNote: (newCurrentNote: any) => void;
    showNoteDetails: any;
    setShowNoteDetails: (showNote: any) => void;

    //data?: AssetGroupTagHistoryRecord[];
}

export const HistoryTableContext = createContext<HistoryTableContext | null>(null);
export const useHistoryTableContext = () => {
    const context = useContext(HistoryTableContext);
    if (!context) {
        throw new Error('useHistoryTableContext is outside of HistoryTableContext');
    }
    return context;
};
