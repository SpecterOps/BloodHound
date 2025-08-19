import { createContext, ReactNode, useContext, useState } from 'react';

export const FileUploadDialogContext = createContext<{
    setShowFileIngestDialog: React.Dispatch<React.SetStateAction<boolean>>;
    showFileIngestDialog: boolean;
}>({
    setShowFileIngestDialog: () => {},
    showFileIngestDialog: false,
});

export const FileUploadDialogProvider = ({ children }: { children: ReactNode }) => {
    const [showFileIngestDialog, setShowFileIngestDialog] = useState(false);

    const value = {
        showFileIngestDialog,
        setShowFileIngestDialog,
    };

    return <FileUploadDialogContext.Provider value={value}>{children}</FileUploadDialogContext.Provider>;
};

export const useFileUploadDialogContext = () => useContext(FileUploadDialogContext);
