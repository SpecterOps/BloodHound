import {
    ActiveDirectoryPathfindingEdges,
    ActiveDirectoryRelationshipKind,
    AzurePathfindingEdges,
    AzureRelationshipKind,
} from '../../../../graphSchema';
import { AllEdgeTypes } from './edgeTypes';

describe('Make sure pathfinding filterable edges match schemagen', () => {
    it('matches', () => {
        const adEdges = getEdgeListFromCategory('Active Directory');
        const azEdges = getEdgeListFromCategory('Azure');

        expect(adEdges).toBeDefined();
        expect(azEdges).toBeDefined();

        if (adEdges && azEdges) {
            const adSchemaEdges = ActiveDirectoryPathfindingEdges();
            const azSchemaEdges = AzurePathfindingEdges();

            const adExclusiveToFilter = [];
            const adExclusiveToSchema = [];

            for (const schemaEdge of adSchemaEdges) {
                if (!adEdges.includes(schemaEdge)) {
                    adExclusiveToSchema.push(schemaEdge);
                }
            }

            for (const edge of adEdges) {
                if (!adSchemaEdges.includes(edge as ActiveDirectoryRelationshipKind)) {
                    adExclusiveToFilter.push(edge);
                }
            }

            const azExclusiveToFilter = [];
            const azExclusiveToSchema = [];

            for (const schemaEdge of azSchemaEdges) {
                if (!azEdges.includes(schemaEdge)) {
                    azExclusiveToSchema.push(schemaEdge);
                }
            }

            for (const edge of azEdges) {
                if (!azSchemaEdges.includes(edge as AzureRelationshipKind)) {
                    azExclusiveToFilter.push(edge);
                }
            }

            console.log({ adExclusiveToFilter, adExclusiveToSchema });

            console.log({ azExclusiveToFilter, azExclusiveToSchema });

            expect(adEdges.length).toEqual(adSchemaEdges.length);
            expect(azEdges.length).toEqual(azSchemaEdges.length);
        }
    });
});

function getEdgeListFromCategory(categoryName: string) {
    const category = AllEdgeTypes.find((category) => category.categoryName === categoryName);
    return category?.subcategories.flatMap((subcategory) => subcategory.edgeTypes);
}
