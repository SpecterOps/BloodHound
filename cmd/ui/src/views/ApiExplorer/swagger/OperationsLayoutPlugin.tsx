// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { Box, Typography, Skeleton, useTheme } from '@mui/material';
import { PageWithTitle } from 'bh-shared-ui';

type Props = {
    getComponent: (
        name: string,
        isContainer?: boolean,
        options?: { failSilently: boolean }
    ) => React.JSXElementConstructor<any>;
    specSelectors: {
        specStr: () => string;
        loadingStatus: () => string;
        isOAS3: () => boolean;
        isSwagger2: () => boolean;
    };
};

function CustomLayout(props: Props) {
    const theme = useTheme();
    const { getComponent, specSelectors } = props;
    const VersionPragmaFilter = getComponent('VersionPragmaFilter', true);
    const FilterContainer = getComponent('FilterContainer', true);
    const Operations = getComponent('operations', true);
    const Models = getComponent('Models', true);
    const Errors = getComponent('errors', true);
    const SvgAssets = getComponent('SvgAssets');

    const isOAS3 = specSelectors.isOAS3();
    const isSwagger2 = specSelectors.isSwagger2();
    const isReady = () => specSelectors.loadingStatus() === 'success';

    return (
        <PageWithTitle title='API Explorer' data-testid='api-explorer'>
            {!isReady() ? (
                <Box display='grid' gap={theme.spacing(4)}>
                    <Box>
                        <Typography variant='h1'>
                            <Skeleton />
                        </Typography>
                    </Box>
                    <Box>
                        <Skeleton variant='rectangular' height={160} />
                    </Box>
                    <Box>
                        <Skeleton variant='rectangular' height={80} />
                    </Box>
                </Box>
            ) : (
                <Box className='swagger-ui' display='grid' gap={theme.spacing(4)}>
                    <SvgAssets />
                    <VersionPragmaFilter isSwagger2={isSwagger2} isOAS3={isOAS3} alsoShow={<Errors />}>
                        <Box>
                            <FilterContainer />
                        </Box>
                        <Box>
                            <Operations />
                        </Box>
                        <Box>
                            <Models />
                        </Box>
                    </VersionPragmaFilter>
                </Box>
            )}
        </PageWithTitle>
    );
}

export const OperationsLayoutPlugin = () => {
    return {
        components: {
            OperationsLayout: CustomLayout,
        },
    };
};
