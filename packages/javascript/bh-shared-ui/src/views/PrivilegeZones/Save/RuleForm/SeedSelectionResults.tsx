import { Card, CardContent, CardHeader } from '@bloodhoundenterprise/doodleui';
import {
    SeedExpansionMethod,
    SeedExpansionMethodAll,
    SeedExpansionMethodChild,
    SeedExpansionMethodNone,
} from 'js-client-library';
import { FC, useContext, useMemo } from 'react';
import { useQuery } from 'react-query';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { useOwnedTagId } from '../../../../hooks'; //specify file
import { usePZPathParams } from '../../../../hooks/usePZParams';
import { apiClient } from '../../../../utils';
import RuleFormContext from './RuleFormContext';

const getRuleExpansionMethod = (
    tagId: string,
    tagType: 'labels' | 'zones',
    ownedId: string | undefined
): SeedExpansionMethod => {
    // Owned is a specific tag type that does not undergo expansion
    if (tagId === ownedId) return SeedExpansionMethodNone;

    return tagType === 'zones' ? SeedExpansionMethodAll : SeedExpansionMethodChild;
};

const EmptySeedResults: FC<{ className: string; displayText: string }> = ({ className, displayText }) => {
    return <p className={className}>{displayText}</p>;
};

export const SeedSelectionResults: FC = () => {
    const { seeds, ruleType } = useContext(RuleFormContext);
    const { tagType, tagId } = usePZPathParams();
    const ownedId = useOwnedTagId();

    const expansion = getRuleExpansionMethod(tagId, tagType, ownedId?.toString());

    const { data: sampleResults, isFetched: sampleResultsFetched } = useQuery({
        queryKey: ['privilege-zones', 'preview-selectors', ruleType, seeds, expansion],
        queryFn: async ({ signal }) => {
            return apiClient
                .assetGroupTagsPreviewSelectors({ seeds, expansion }, { signal })
                .then((res) => res.data.data['members']);
        },
        retry: false,
        refetchOnWindowFocus: false,
        enabled: seeds.length > 0,
    });

    const directObjects = useMemo(
        () => sampleResults?.filter((objectItem) => objectItem.source === 1),
        [sampleResults]
    );
    const expandedObjects = useMemo(
        () => sampleResults?.filter((objectItem) => objectItem.source > 1),
        [sampleResults]
    );

    const setRuleTypeDisplay = () => {
        switch (ruleType) {
            case 1:
                return 'Object';
            case 2:
                return 'Cypher';
            default:
                return '';
        }
    };
    return (
        <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
            <CardHeader className='pl-6 first:py-6 text-xl font-bold'>Sample Results</CardHeader>
            {sampleResultsFetched ? (
                <>
                    <CardContent className='pl-4'>
                        <div className='font-bold pl-2 mb-2'>Direct Objects</div>
                        {directObjects?.length ? (
                            <VirtualizedNodeList nodes={directObjects} itemSize={46} heightScalar={10} />
                        ) : (
                            <EmptySeedResults className='pl-2' displayText='No Direct Objects found' />
                        )}
                    </CardContent>
                    <CardContent className='pl-4'>
                        <div className='font-bold pl-2 mb-2'>Expanded Objects</div>
                        {expandedObjects?.length ? (
                            <VirtualizedNodeList nodes={expandedObjects} itemSize={46} heightScalar={10} />
                        ) : (
                            <EmptySeedResults className='pl-2' displayText='No Expanded Objects found' />
                        )}
                    </CardContent>
                </>
            ) : (
                <EmptySeedResults
                    className='pl-6'
                    displayText={`Enter ${setRuleTypeDisplay()} Rule form information to see sample results`}
                />
            )}
        </Card>
    );
};
