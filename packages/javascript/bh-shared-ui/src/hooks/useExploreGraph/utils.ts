import { AllEdgeTypes, Category, EdgeCheckboxType, Subcategory } from '../../edgeTypes';

export const extractEdgeTypes = (edges: EdgeCheckboxType[]): string[] => {
    return edges.filter((edge) => edge.checked).map((edge) => edge.edgeType);
};

export const mapParamsToFilters = (params: string[], initial: EdgeCheckboxType[]): EdgeCheckboxType[] => {
    return initial.map((edge) => ({
        ...edge,
        checked: !!params.includes(edge.edgeType),
    }));
};

export const compareEdgeTypes = (initial: string[], comparison: string[]): boolean => {
    const a = initial.slice(0).sort();
    const b = comparison.slice(0).sort();

    return a.length === b.length && a.every((item, index) => item === b[index]);
};

// Create a list of all edge types to initialize pathfinding filter state
export const getInitialPathFilters = (): EdgeCheckboxType[] => {
    const initialPathFilters: EdgeCheckboxType[] = [];

    AllEdgeTypes.forEach((category: Category) => {
        category.subcategories.forEach((subcategory: Subcategory) => {
            subcategory.edgeTypes.forEach((edgeType: string) => {
                initialPathFilters.push({
                    category: category.categoryName,
                    subcategory: subcategory.name,
                    edgeType,
                    checked: true,
                });
            });
        });
    });

    return initialPathFilters;
};
