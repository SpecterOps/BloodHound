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

import { Alert, AlertTitle, Box, useTheme } from '@mui/material';

type NoDataAlertProps = {
    dataCollectionLink: JSX.Element;
    fileIngestLink?: JSX.Element;
    sampleDataLink?: JSX.Element;
};

export const NoDataAlert: React.FC<NoDataAlertProps> = ({ dataCollectionLink, fileIngestLink, sampleDataLink }) => {
    const theme = useTheme();

    return (
        <Box display={'flex'} justifyContent={'center'} mt={theme.spacing(8)} mx={theme.spacing(4)}>
            <Alert severity={'info'}>
                <AlertTitle>No Data Available</AlertTitle>
                <p>
                    It appears that no data has been uploaded yet. See our {dataCollectionLink} documentation to learn
                    how to start collecting data.
                </p>

                {fileIngestLink && (
                    <p>
                        If you have files available from a SharpHound or AzureHound collection, please visit the{' '}
                        {fileIngestLink} page to begin uploading your data.
                    </p>
                )}

                {sampleDataLink && (
                    <p>
                        If you want to test BloodHound with sample data, you may download some from our {sampleDataLink}{' '}
                        GitHub page.
                    </p>
                )}
            </Alert>
        </Box>
    );
};
