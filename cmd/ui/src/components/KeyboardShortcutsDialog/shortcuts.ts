export type ShortCutsMap = Record<string, string[][]>;

export const EXPLORE_SHORTCUTS = {
    'Explore Page': [
        ['S', 'Jump to Node Search'],
        ['P', 'Jump to Pathfinding'],
        ['C', 'Focus Cypher Query Editor'],
        ['Shift + S', 'Save Current Query'],
        ['R', 'Run Current Cypher Query'],
        ['/', 'Search Current Nodes'],
        ['T', 'Toggle Table View'],
        ['I', 'Toggle Node Info Panel'],
        ['Shift + G', 'Reset Graph View'],
    ],
};

export const GLOBAL_SHORTCUTS = {
    Global: [
        ['<integer>', 'Navigate sidebar pages'],
        ['H', 'Launch keyboard shortcuts list'],
        ['U', 'Launch File Upload dialog'],
        ['M', 'Toggle Dark Mode'],
    ],
};

export const POSTURE_PAGE_SHORTCUTS = {
    'Posture Page': [['F', 'Filter Table Data']],
};
