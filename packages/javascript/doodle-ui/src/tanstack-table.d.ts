import '@tanstack/react-table';
import type { RowData } from '@tanstack/react-table';

declare module '@tanstack/react-table' {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    interface ColumnMeta<TData extends RowData, TValue> {
        label?: string; // Add your custom property here, making it optional with '?' if needed
        // You can add other custom properties here as well
        // exampleProp?: boolean;
    }
}
