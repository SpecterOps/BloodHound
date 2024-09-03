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
    ReconciliationToggle = 'analysis.reconciliation',
    ReconcilationThreshold = 'analysis.reconciliation_threshold',
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

export type ReconciliationToggleConfiguration = {
    key: ConfigurationKey.ReconciliationToggle;
    value: {
        enabled: boolean;
    };
};

export type ReconciliationThresholdConfiguration = {
    key: ConfigurationKey.ReconcilationThreshold;
    value: {
        session_threshold: string;
        general_threshold: string;
    };
};

export type ConfigurationPayload =
    | PasswordExpirationConfiguration
    | Neo4jConfiguration
    | CitrixConfiguration
    | ReconciliationToggleConfiguration
    | ReconciliationThresholdConfiguration;

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

export const parseReconciliationToggleConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<ReconciliationToggleConfiguration> | undefined => {
    const key = ConfigurationKey.ReconciliationToggle;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};

export const parseReconciliationThresholdConfiguration = (
    response: GetConfigurationResponse | undefined
): ConfigurationWithMetadata<ReconciliationThresholdConfiguration> | undefined => {
    const key = ConfigurationKey.ReconcilationThreshold;
    const config = getConfigurationFromKey(response, key);

    return config?.key === key ? config : undefined;
};
