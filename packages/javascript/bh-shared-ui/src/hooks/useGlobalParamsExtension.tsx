import { EnvironmentQueryParams, useEnvironmentParams } from './useEnvironmentParams';
import { PZQueryParams, usePZQueryParams } from './usePZParams';

type GlobalParams = EnvironmentQueryParams & PZQueryParams;

type GloballySupportedParamKeys = keyof GlobalParams;

export type GlobalParamsExtension<T> = GlobalParams & T;

export const GloballySupportedSearchParams = [
    'environmentId',
    'environmentAggregation',
    'assetGroupTagId',
] satisfies GloballySupportedParamKeys[];

// Params pulled into this hook will be plumbed into each page level param hook.
export const useGlobalParamsExtension = (): GlobalParams => {
    const { setEnvironmentParams, ...environmentRest } = useEnvironmentParams();
    const { assetGroupTagId } = usePZQueryParams();
    return { ...environmentRest, assetGroupTagId };
};
