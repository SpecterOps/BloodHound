import { atom } from 'jotai';
import { GraphData } from 'js-client-library';

export const isDialogOpenAtom = atom(false);

export const isShiftDownAtom = atom(false);

export const getEdgePayload = (data: GraphData, source: string, target: string, edgeKind: string) => {
    return {
        source_object_id: data.nodes[source].objectId,
        target_object_id: data.nodes[target].objectId,
        edge_kind: edgeKind,
        properties: {},
    };
};
