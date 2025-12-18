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
        ['H', 'View keyboard shortcuts dialog'],
        ['M', 'Toggle Dark Mode'],
        ['<integer>', 'Navigate sidebar pages'],
    ],
};

export const ATTACK_PATHS_SHORTCUTS = {
    'Attack Paths Page': [
        ['Shift + R', 'Reset to Default View'],
        ['.', 'Jump to Next Finding'],
        [',', 'Jump to Previous Finding'],
        ['E', 'Jump to Environment Selector'],
    ],
};

export const POSTURE_PAGE_SHORTCUTS = {
    'Posture Page': [
        ['E', 'Jump to Environment Selector'],
        ['Z', 'Jump to Zone Selector'],
        ['F', 'Filter Table Data'],
    ],
};
