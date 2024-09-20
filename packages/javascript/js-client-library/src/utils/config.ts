import { ConfigurationWithMetadata, GetConfigurationResponse } from '../responses';

/* 
A collection of types and helper functions for working with values from the config endpoint.

Intended Usage:

const { data } = useGetConfiguration();                   // get data from query as <GetConfigurationResponse | undefined>
const neo4jConfig = parseNeo4jConfiguration(data);        // narrow to <ConfigurationWithMetadata<Neo4jConfiguration> | undefined>

if (neo4jConfig) {
    console.log(neo4jConfig.key)                          // ok
    console.log(neo4jConfig.value.batch_write_size)       // ok
    console.log(neo4jConfig.value.enabled)                // type error
}
*/
export enum ConfigurationKey {
    PasswordExpiration = 'auth.password_expiration_window',
    Neo4j = 'neo4j.configuration',
    Citrix = 'analysis.citrix_rdp_support',
    Reconciliation = 'analysis.reconciliation',
    PruneTTL = 'prune.ttl',
}

export type PasswordExpirationConfiguration = {
    key: ConfigurationKey.PasswordExpiration;
    value: {
        duration: string;
    };
};

export type Neo4jConfiguration = {
    key: ConfigurationKey.Neo4j;
    value: {
        batch_write_size: number;
        write_flush_size: number;
    };
};

export type CitrixConfiguration = {
    key: ConfigurationKey.Citrix;
    value: {
        enabled: boolean;
    };
};

export type ReconciliationConfiguration = {
    key: ConfigurationKey.Reconciliation;
    value: {
        enabled: boolean;
    };
};

export type PruneTTLConfiguration = {
    key: ConfigurationKey.PruneTTL;
    value: {
        has_session_edge_ttl: string;
        base_ttl: string;
    };
};

export type ConfigurationPayload =
    | PasswordExpirationConfiguration
    | Neo4jConfiguration
    | CitrixConfiguration
    | ReconciliationConfiguration
    | PruneTTLConfiguration;

export const getConfigurationFromKey = (config: GetConfigurationResponse | undefined, key: ConfigurationKey) => {
    return config?.data.find((c) => c.key === key);
};

export const parsePasswordExpirationConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<PasswordExpirationConfiguration> | undefined => {
    const key = ConfigurationKey.PasswordExpiration;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};

export const parseNeo4jConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<Neo4jConfiguration> | undefined => {
    const key = ConfigurationKey.Neo4j;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};

export const parseCitrixConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<CitrixConfiguration> | undefined => {
    const key = ConfigurationKey.Citrix;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};

export const parseReconciliationConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<ReconciliationConfiguration> | undefined => {
    const key = ConfigurationKey.Reconciliation;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};

export const parsePruneTTLConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<PruneTTLConfiguration> | undefined => {
    const key = ConfigurationKey.PruneTTL;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};
