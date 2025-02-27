import { useAvailableEnvironments, useEnvironmentParams, useFeatureFlag, useMatchingPaths } from 'bh-shared-ui';
import { Domain } from 'js-client-library';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { setDomain } from 'src/ducks/global/actions';
import { useAppDispatch, useAppSelector } from 'src/store';

// Future Dev: when we implement deep linking support for selected domain in BHE, move this to shared-ui and rip out the reducer logic (including stateUpdater)
const useInitialEnvironment = (envSupportedRoutes: string[]) => {
    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const { data: flag } = useFeatureFlag('back_button_support', {
        enabled: !!authState.isInitialized && fullyAuthenticated,
    });

    const reduxEnvironment = useAppSelector((state) => state.global.options.domain);
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const dispatch = useAppDispatch();

    const { environmentId, setEnvironmentParams } = useEnvironmentParams();
    const environmentSupportedRoute = useMatchingPaths(envSupportedRoutes);

    const currentEnvironmentId = flag?.enabled ? environmentId : reduxEnvironment?.id;

    const stateUpdater = (environment: Domain | null) => {
        if (flag?.enabled) {
            setEnvironmentParams({ environmentId: environment?.id });
        } else {
            dispatch(setDomain(environment));
        }
    };

    useAvailableEnvironments({
        appendQueryKey: ['initial-environment'],
        // set initial environment/tenant once user is authenticated
        enabled: isFullyAuthenticated && environmentSupportedRoute,
        onError: () => stateUpdater(null),
        onSuccess: (availableEnvironments) => {
            if (!availableEnvironments?.length || currentEnvironmentId) return;

            const collectedEnvironments = availableEnvironments
                ?.filter((environment: Domain) => environment.collected) // omit uncollected environments
                .sort((a: Domain, b: Domain) => b.impactValue - a.impactValue); // sort by impactValue descending

            if (collectedEnvironments?.length) {
                stateUpdater(collectedEnvironments[0]);
            } else {
                stateUpdater(null);
            }
        },
    });
};

export default useInitialEnvironment;
