import { findIconDefinition, IconName } from '@fortawesome/fontawesome-svg-core';
import { apiClient, GenericQueryOptions, IconDictionary } from 'bh-shared-ui';
import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { appendSvgUrls, DEFAULT_ICON_COLOR, NODE_SCALE } from 'src/views/Explore/svgIcons';

export const getCustomNodeKinds = async (options: RequestOptions): Promise<IconDictionary> =>
    apiClient.getCustomNodeKinds(options).then((res) => {
        const customIcons: IconDictionary = {};

        res.data.data.forEach((node) => {
            const iconName = node.config.icon.name as IconName;

            const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: iconName });
            if (iconDefinition == undefined) {
                return;
            }

            customIcons[node.kindName] = {
                icon: iconDefinition,
                color: DEFAULT_ICON_COLOR,
            };
        });

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

export function transformIconDictionary(customIcons: IconDictionary): IconDictionary {
    appendSvgUrls(customIcons, NODE_SCALE);

    return customIcons;
}
