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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Alert, AlertTitle } from '@mui/material';
import { useEffect } from 'react';
import { useNavigate } from 'react-router';
import LoginPage from 'src/components/LoginPage';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { ROUTE_EXPLORE, ROUTE_LOGIN } from 'src/routes/constants';
import { useAppSelector } from 'src/store';

const NotFound: React.FC = () => {
    const navigate = useNavigate();
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);

    // Redirect to login if unauthenticated
    useEffect(() => {
        if (isFullyAuthenticated) {
            return;
        }

        navigate(ROUTE_LOGIN);
    }, [isFullyAuthenticated, navigate]);

    return (
        <LoginPage>
            <div className='flex flex-col gap-6'>
                <Alert severity='warning'>
                    <AlertTitle>404 - Page not found</AlertTitle>
                    There is no page associated with this route. Please contact your system administrator for
                    assistance.
                </Alert>

                <Button
                    onClick={() => {
                        navigate(ROUTE_EXPLORE);
                    }}
                    data-testid='page-not-found-go-to-explore'
                    size='large'
                    type='button'
                    className='w-full'>
                    Go to Explore
                </Button>
            </div>
        </LoginPage>
    );
};

export default NotFound;
