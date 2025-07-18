import { FlatGraphResponse, GraphNodeSpreadWithProperties } from 'js-client-library';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';

export const makeStoreMapFromColumnOptions = (columnOptions: ManageColumnsComboBoxOption[]) =>
    columnOptions.reduce(
        (acc, col) => {
            acc[col?.id] = true;

            return acc;
        },
        {} as Record<string, boolean>
    );

export type NodeClickInfo = { id: string; x: number; y: number };
export type MungedTableRowWithId = GraphNodeSpreadWithProperties & { id: string };

const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'objectid', 'name', 'isTierZero'];

export const requiredColumns = Object.fromEntries(REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.map((key) => [key, true]));

export const isSmallColumn = (key: string, value: any) => key === 'nodetype' || typeof value === 'boolean';

export const isIconField = (value: any) => typeof value === 'boolean' || value === undefined || value === null;

export interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    data?: FlatGraphResponse;
    selectedColumns?: Record<string, boolean>;
    onRowClick?: (row: MungedTableRowWithId) => void;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    selectedNode: string | null;
    onDownloadClick: () => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
}
