// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { faPlay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { 
    Box, 
    Button, 
    Typography,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Paper,
    CircularProgress,
    Alert
} from '@mui/material';
import {
    PageWithTitle,
    apiClient,
    encodeCypherQuery,
    ExploreQueryParams,
    createTypedSearchParams,
    useAppNavigate,
} from 'bh-shared-ui';
import { useState } from 'react';
import { ROUTE_EXPLORE } from 'src/routes/constants';

interface QueryResults {
    startingNodes: Array<{ id: string; name: string; kind: string; objectId: string }>;
}

const SniffDeep = () => {
    const [isLoading, setIsLoading] = useState(false);
    const [results, setResults] = useState<QueryResults | null>(null);
    const [error, setError] = useState<string | null>(null);
    const navigate = useAppNavigate();

    const cypherQuery = `
        MATCH p1 = (n:Base)-[:GenericAll|AddMember|MemberOf*..]->(:Group)-[:GetChanges]->(d:Domain) 
        WHERE NOT (n)-[:GenericAll|AddMember|MemberOf*0..]->()-[:DCSync]->(d) 
        MATCH p2 = (n:Base)-[:GenericAll|AddMember|MemberOf*..]->(:Group)-[:GetChangesAll]->(d:Domain) 
        RETURN n
    `;

    const handleExploreNode = (objectId: string) => {
        const exploreQuery = `MATCH p1 = (n:Base)-[:GenericAll|AddMember|MemberOf*..]->(:Group)-[:GetChanges]->(d:Domain)
WHERE n.objectid = "${objectId}"
MATCH p2 = (n:Base)-[:GenericAll|AddMember|MemberOf*..]->(:Group)-[:GetChangesAll]->(d:Domain)
RETURN p1,p2`;
        
        navigate({
            pathname: ROUTE_EXPLORE,
            search: createTypedSearchParams<ExploreQueryParams>({
                searchType: 'cypher',
                cypherSearch: encodeCypherQuery(exploreQuery),
                exploreSearchTab: 'cypher',
            }),
        });
    };

    const handlePlayClick = async () => {
        setIsLoading(true);
        setError(null);
        setResults(null);

        try {
            const response = await apiClient.cypherSearch(cypherQuery, { signal: new AbortController().signal }, true);
            
            if (response.data && response.data.data) {
                const graphData = response.data.data;
                
                // Extract starting nodes (n) from the query results
                const startingNodes = new Map();

                // Since we're returning only 'n' nodes, all returned nodes should be starting nodes
                Object.entries(graphData.nodes || {}).forEach(([nodeId, nodeData]) => {
                    startingNodes.set(nodeId, {
                        id: nodeId,
                        name: nodeData.label || nodeData.objectId || nodeId,
                        kind: nodeData.kind,
                        objectId: nodeData.objectId || nodeId
                    });
                });

                console.log('Debug - All nodes:', Object.entries(graphData.nodes || {}).map(([id, node]) => ({
                    id, 
                    kind: node.kind, 
                    label: node.label,
                    objectId: node.objectId
                })));
                console.log('Debug - Starting nodes found:', Array.from(startingNodes.values()));

                setResults({
                    startingNodes: Array.from(startingNodes.values())
                });
            } else {
                setError('No data returned from query');
            }
        } catch (err: any) {
            setError(err.message || 'Failed to execute query');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <PageWithTitle title="Sniff Deep" data-testid="sniff-deep-page">
            <Box sx={{ p: 3 }}>
                {/* Header Section */}
                <Box sx={{ 
                    display: 'flex', 
                    flexDirection: 'column', 
                    alignItems: 'center',
                    mb: 4
                }}>
                    <Typography variant="h4" component="h1" gutterBottom>
                        Sniff Deep Analysis
                    </Typography>
                    <Typography variant="body1" color="text.secondary" gutterBottom sx={{ mb: 3 }}>
                        Execute Cypher query to find paths with DCSync-like privileges
                    </Typography>
                    
                    <Button
                        variant="contained"
                        color="primary"
                        size="large"
                        startIcon={isLoading ? <CircularProgress size={20} color="inherit" /> : <FontAwesomeIcon icon={faPlay} />}
                        data-testid="sniff-deep-play-button"
                        onClick={handlePlayClick}
                        disabled={isLoading}
                    >
                        {isLoading ? 'Running Query...' : 'Start Sniffing'}
                    </Button>
                </Box>

                {/* Error Display */}
                {error && (
                    <Alert severity="error" sx={{ mb: 3 }}>
                        {error}
                    </Alert>
                )}

                {/* Results Section */}
                {results && (
                    <Box>
                        <Typography variant="h5" gutterBottom sx={{ mb: 3 }}>
                            Query Results
                        </Typography>
                        
                        {/* Starting Nodes Table */}
                        <Box sx={{ maxWidth: '1200px' }}>
                            <Typography variant="h6" gutterBottom>
                                Attack Starting Points - {results.startingNodes.length}
                            </Typography>
                            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                                Users/Computers that can initiate DCSync-like attack paths
                            </Typography>
                            <TableContainer component={Paper} sx={{ maxHeight: 600 }}>
                                <Table stickyHeader size="small">
                                    <TableHead>
                                        <TableRow>
                                            <TableCell>Name</TableCell>
                                            <TableCell>Kind</TableCell>
                                            <TableCell>Object ID</TableCell>
                                            <TableCell>Actions</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {results.startingNodes.map((node, index) => (
                                            <TableRow key={index}>
                                                <TableCell>{node.name}</TableCell>
                                                <TableCell>{node.kind}</TableCell>
                                                <TableCell sx={{ 
                                                    fontSize: '0.75rem', 
                                                    fontFamily: 'monospace',
                                                    maxWidth: '200px',
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis'
                                                }}>
                                                    {node.objectId}
                                                </TableCell>
                                                <TableCell>
                                                    <Button
                                                        variant="outlined"
                                                        size="small"
                                                        onClick={() => handleExploreNode(node.objectId)}
                                                    >
                                                        Explore Paths
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                        {results.startingNodes.length === 0 && (
                                            <TableRow>
                                                <TableCell colSpan={4} sx={{ textAlign: 'center', fontStyle: 'italic' }}>
                                                    No attack starting points found
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </Box>
                    </Box>
                )}
            </Box>
        </PageWithTitle>
    );
};

export default SniffDeep;
