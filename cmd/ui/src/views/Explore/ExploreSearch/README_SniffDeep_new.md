# SniffDeep Search Component

## Overview

The SniffDeepSearch component provides a specialized search interface for discovering specific DAWGS paths related to DCSync and other Active Directory attacks. Unlike the standard pathfinding search, it focuses on specific edge types from Group nodes to destination nodes.

## Features

### Custom Search Interface
- **Fixed Source**: Always searches from Group nodes (displayed as "Group nodes (source)")
- **Dynamic Destination**: Users can search for and select any destination node
- **Specialized Queries**: Generates specific DAWGS queries for GetChanges and GetChangesAll edges

### Search Options
- **All**: Searches for both GetChanges and GetChangesAll edges
- **DCSync**: Focuses on DCSync-related paths (currently includes both edge types)

### UI Components
- **Dropdown Selector**: Positioned above the search area for selecting search type
- **Source Field**: Read-only field showing "Group nodes (source)" 
- **Destination Field**: Functional search combobox for selecting target nodes
- **Play Button**: Manual trigger for executing the search (disabled until destination is selected)
- **Filter Button**: Placeholder (disabled) to maintain visual consistency

## DAWGS Query Generation

The component generates specific Cypher queries based on the selected search type:

### Path Type 1: GetChanges Edge
```cypher
MATCH p=(g:Group)-[:GetChanges]->(d)
WHERE d.objectid = "{destinationNodeId}"
RETURN p
LIMIT 1000
```

### Path Type 2: GetChangesAll Edge  
```cypher
MATCH p=(g:Group)-[:GetChangesAll]->(d)
WHERE d.objectid = "{destinationNodeId}"
RETURN p
LIMIT 1000
```

## Implementation Details

### Custom Hooks
- `useSniffDeepSearch`: Manages destination node search state independently from pathfinding
- Handles node selection, search term management, and state updates

### Search Logic
- Validates destination node selection before executing queries
- Logs generated queries for debugging/monitoring
- TODO: Integration with actual cypher search API and graph visualization

### Visual Design
- Maintains PathfindingSearch appearance for consistency
- Gray container background with rounded corners
- Proper spacing and alignment with existing UI elements

## Usage

```tsx
const SniffDeepSearch = ({
    pathfindingSearchState,
    pathfindingFilterState,
}: {
    pathfindingSearchState: ReturnType<typeof usePathfindingSearch>;
    pathfindingFilterState: ReturnType<typeof usePathfindingFilters>;
}) => {
    // Component implementation
};
```

## Testing

The component includes comprehensive tests covering:
- Dropdown option selection (All/DCSync)
- Destination node search functionality  
- Play button enable/disable state
- Query generation logic
- UI element rendering and interaction

## Future Enhancements

1. **API Integration**: Connect generated queries to actual cypher search endpoint
2. **Graph Visualization**: Display results in the explore graph view
3. **Additional Path Types**: Extend beyond GetChanges/GetChangesAll edges
4. **Advanced Filtering**: Add support for additional search criteria
5. **Performance Optimization**: Implement query caching and pagination

## Dependencies

- `bh-shared-ui`: ExploreSearchCombobox, DropdownSelector, hooks
- `@bloodhoundenterprise/doodleui`: Button component
- `@fortawesome/react-fontawesome`: Icons for UI elements
