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

import React, { useEffect, useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { clearLoginRedirect, getLoginDestination } from 'src/routes/loginRedirect';

const Home: React.FC = () => {
    const { state } = useLocation();
    // AuthenticatedRoute mounts Home only after login and any required password reset.
    // Capture the destination before clearing storage so rerenders keep the same target.
    const [redirectTo] = useState(() => getLoginDestination(state));

    useEffect(() => {
        clearLoginRedirect();
    }, []);

    return <Navigate to={redirectTo} replace />;
};

export default Home;
