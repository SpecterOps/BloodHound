import { findIconDefinition, IconName } from '@fortawesome/fontawesome-svg-core';
import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { apiClient, DEFAULT_ICON_BACKGROUND, GenericQueryOptions, IconDictionary } from '../utils';

export const getCustomNodeKinds = async (options: RequestOptions): Promise<IconDictionary> =>
    apiClient.getCustomNodeKinds(options).then((res) => {
        const customIcons: IconDictionary = {};

        if (Array.isArray(res?.data?.data)) {
            res.data.data.forEach((node) => {
                const iconName = node.config.icon.name as IconName;

                const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: iconName });
                if (iconDefinition == undefined) {
                    return;
                }

                customIcons[node.kindName] = {
                    icon: iconDefinition,
                    color: DEFAULT_ICON_BACKGROUND,
                };
            });
        }

        return customIcons;
    });

export const useCustomNodeKinds = (
    queryOptions?: GenericQueryOptions<IconDictionary>
): UseQueryResult<IconDictionary> => {
    return useQuery({
        queryKey: ['getCustomNodeKinds'],
        queryFn: ({ signal }) => getCustomNodeKinds({ signal }),
        staleTime: 2 * (60 * 1000),
        cacheTime: 5 * (60 * 1000),
        ...queryOptions,
    });
};
