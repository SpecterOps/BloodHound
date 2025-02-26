import { useEnvironment, useEnvironmentParams, useFeatureFlag } from 'bh-shared-ui';
import { useAppSelector } from 'src/store';

const useGroupManagementStateSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');
    const { environmentId } = useEnvironmentParams();
    const { data: environmentFromParams } = useEnvironment(environmentId, { enabled: flag?.enabled });

    const globalDomain = useAppSelector((state) => state.global.options.domain);

    return flag?.enabled ? environmentFromParams ?? null : globalDomain;
};

export default useGroupManagementStateSwitch;
