import { FlatGraphResponse, GraphNodeSpreadWithProperties, GraphResponse } from 'js-client-library';
import { isGraphResponse } from '../..';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';

export const makeStoreMapFromColumnOptions = (columnOptions: ManageColumnsComboBoxOption[]) =>
    columnOptions.reduce(
        (acc, col) => {
            acc[col?.id] = true;

            return acc;
        },
        {} as Record<string, boolean>
    );

const KNOWN_NODE_KEYS = ['kind', 'objectId', 'label', 'isTierZero'];

export const getExploreTableData = (graphData: GraphResponse | FlatGraphResponse | undefined) => {
    if (!graphData || !isGraphResponse(graphData)) return;

    const nodes = graphData.data.nodes;
    const unknownNodeKeys = graphData.data.node_keys;

    const completeNodeKeys = unknownNodeKeys?.concat(KNOWN_NODE_KEYS);

    return {
        nodes,
        node_keys: completeNodeKeys,
    };
};

export type NodeClickInfo = { id: string; x: number; y: number };
export type MungedTableRowWithId = GraphNodeSpreadWithProperties & { id: string };

export const knownColumns = Object.fromEntries(KNOWN_NODE_KEYS.map((key) => [key, true]));

export interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    selectedColumns?: Record<string, boolean>;
    onRowClick?: (row: MungedTableRowWithId) => void;
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    selectedNode: string | null;
    onDownloadClick: () => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
}
