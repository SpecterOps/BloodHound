import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';

export const makeStoreMapFromColumnOptions = (columnOptions: ManageColumnsComboBoxOption[]) =>
    columnOptions.reduce(
        (acc, col) => {
            acc[col?.id] = true;

            return acc;
        },
        {} as Record<string, boolean>
    );
