import { FileUploadDialogProvider, KeyboardShortcutsDialogProvider } from 'bh-shared-ui';
import { ReactNode } from 'react';

const DialogProviders = ({ children }: { children: ReactNode }) => (
    <FileUploadDialogProvider>
        <KeyboardShortcutsDialogProvider>{children}</KeyboardShortcutsDialogProvider>
    </FileUploadDialogProvider>
);

export default DialogProviders;
