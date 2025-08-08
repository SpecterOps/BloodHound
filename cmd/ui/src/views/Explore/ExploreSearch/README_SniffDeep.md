# Sniff Deep Search Enhancement

## Summary
Added a dropdown menu to the Sniff Deep search tab with "All" and "DCSync" options as requested.

## Changes Made

### 1. Created SniffDeepSearch Component (`/cmd/ui/src/views/Explore/ExploreSearch/SniffDeepSearch.tsx`)
- New component that extends the existing PathfindingSearch functionality
- Includes a dropdown selector with "All" and "DCSync" options
- Maintains the same search interface as pathfinding (start node, destination node, swap button, edge filters)
- Uses the existing DropdownSelector component from bh-shared-ui

### 2. Updated ExploreSearch Component (`/cmd/ui/src/views/Explore/ExploreSearch/ExploreSearch.tsx`)
- Added import for the new SniffDeepSearch component
- Updated the tabs array to use SniffDeepSearch instead of PathfindingSearch for the sniff deep tab
- Maintains all existing functionality for tab switching and parameter handling

### 3. Added Test Coverage (`/cmd/ui/src/views/Explore/ExploreSearch/SniffDeepSearch.test.tsx`)
- Basic test coverage for the new dropdown functionality
- Tests dropdown rendering, option selection, and integration with pathfinding components

## Features Added
1. **Dropdown Menu**: Located at the top of the sniff deep search tab
2. **Two Options Available**:
   - "All" (default selection)
   - "DCSync"
3. **Seamless Integration**: The dropdown doesn't interfere with existing pathfinding search functionality
4. **Future Extensibility**: Easy to add more filter options or integrate with actual search logic

## Technical Notes
- Reuses existing PathfindingSearch component for search functionality
- Uses the same state management hooks (`usePathfindingSearch`, `usePathfindingFilters`)
- Follows existing component patterns and styling
- Ready for integration with backend filtering logic when DCSync functionality is implemented

## Next Steps
The dropdown is currently functional for UI purposes. To complete the feature:
1. Connect the dropdown selection to actual search filtering logic
2. Implement DCSync-specific search parameters
3. Add any additional dropdown options as needed
