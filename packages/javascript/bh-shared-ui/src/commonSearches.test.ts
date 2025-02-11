import { CommonSearches, CommonSearchType } from './commonSearches';
import {
    ActiveDirectoryNodeKind,
    ActiveDirectoryRelationshipKind,
    AzureNodeKind,
    AzureRelationshipKind,
} from './graphSchema';

describe('common search list', () => {
    const nodeKindPattern = /((?<=\(.?:).*?(?=\)))/gm;
    const relKindPattern = /((?<=\[.?:).*?(?=\]))/gm;

    test('the queries in the list only include nodes and edges that are defined in our schema', () => {
        CommonSearches.forEach((commonSearchType: CommonSearchType) => {
            commonSearchType.queries.forEach((query) => {
                const queryNodeLabels = query.cypher.match(nodeKindPattern);
                const queryRelLabels = query.cypher.match(relKindPattern);

                if (queryNodeLabels) {
                    queryNodeLabels.forEach((result) => {
                        const inAD = Object.values(ActiveDirectoryNodeKind).includes(result as ActiveDirectoryNodeKind);
                        const inAZ = Object.values(AzureNodeKind).includes(result as AzureNodeKind);
                        const inSchema = inAD || inAZ;

                        expect(inSchema).toBeTruthy();
                    });
                }

                if (queryRelLabels) {
                    queryRelLabels.forEach((result) => {
                        // Trim off any depth specifications
                        if (result.includes('*')) result = result.slice(0, result.indexOf('*'));

                        // Turn the match into an array because sometimes there are multiple edges
                        const results = result.split('|');

                        results.forEach((edgeKind) => {
                            const inAD = Object.values(ActiveDirectoryRelationshipKind).includes(
                                edgeKind as ActiveDirectoryRelationshipKind
                            );
                            const inAZ = Object.values(AzureRelationshipKind).includes(
                                edgeKind as AzureRelationshipKind
                            );
                            const inSchema = inAD || inAZ;

                            expect(inSchema).toBeTruthy();
                        });
                    });
                }
            });
        });
    });

    test('typos will be flagged', () => {
        // The typo is 'AZAutmomationContributor'
        const query = `MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)-[:AZOwner|AZUserAccessAdministrator|AZGetCertificates|AZGetKeys|AZGetSecrets|AZAvereContributor|AZKeyVaultContributor|AZContributor|AZVMAdminLogin|AZVMContributor|AZAKSContributor|AZAutmomationContributor|AZLogicAppContributor|AZWebsiteContributor]->(:AZBase)\nRETURN p\nLIMIT 1000`;

        const queryRelLabels = query.match(relKindPattern);

        let hasTypo = false;

        if (queryRelLabels) {
            queryRelLabels.forEach((result) => {
                if (result.includes('*')) result = result.slice(0, result.indexOf('*'));

                const results = result.split('|');

                results.forEach((edgeKind) => {
                    const inAD = Object.values(ActiveDirectoryRelationshipKind).includes(
                        edgeKind as ActiveDirectoryRelationshipKind
                    );
                    const inAZ = Object.values(AzureRelationshipKind).includes(edgeKind as AzureRelationshipKind);
                    const inSchema = inAD || inAZ;

                    // This gets set to true when an edge kind is not in our schema
                    // In this case it is because there is a typo
                    if (!inSchema) hasTypo = true;

                    // The previous test has an assertion at this point and so if a query has a typo
                    // like this test query does the first test block will fail
                });
            });
        }

        expect(hasTypo).toBe(true);
    });
});
