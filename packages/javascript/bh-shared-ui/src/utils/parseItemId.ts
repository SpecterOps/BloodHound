export const parseItemId = (itemId: string) => {
    // Edge identifiers can be either `rel_<sourceNodeId>_<edgeKind>_<targetNodeId>`...
    let match = itemId.match(/^(?:rel_)?(\d+)_(.+)_(\d+)$/);
    if (match) {
        return {
            itemType: 'edge',
            cypherQuery: `MATCH (s)-[r:${match[2]}]->(t) WHERE ID(s) = ${match[1]} AND ID(t) = ${match[3]}  RETURN r LIMIT 1`,
        };
    }

    // or `rel_<edgeId>`...
    match = itemId.match(/^rel_(\d+)$/);
    if (match) {
        return {
            itemType: 'edge',
            cypherQuery: `MATCH ()-[r]->() WHERE ID(r) = ${match[1]} RETURN r LIMIT 1`,
        };
    }

    // otherwise it is a node identifier
    return { itemType: 'node', cypherQuery: `MATCH (n) where ID(n) = ${itemId} RETURN n LIMIT 1` };
};
