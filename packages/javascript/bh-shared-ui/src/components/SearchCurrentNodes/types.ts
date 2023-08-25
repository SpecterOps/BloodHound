export type FlatNode = GraphNode & { id: string };

export type GraphNode = {
    label: string;
    kind: string;
    objectId: string;
    lastSeen: string;
    isTierZero: boolean;
    descendent_count?: number | null;
};

export type GraphNodes = Record<string, GraphNode>;
