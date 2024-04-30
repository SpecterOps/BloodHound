// Copyright 2024 Specter Ops, Inc.
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

import React from 'react';
import {
  useFlag,
  useUnleashClient,
  useUnleashContext,
  useVariant,
  useFlagsStatus,
} from '@unleash/proxy-client-react';

interface IUnleashClassFlagProvider {
  render: (props: any) => React.ReactNode;
  flagName: string;
}

export const UnleashClassFlagProvider = ({
  render,
  flagName,
}: IUnleashClassFlagProvider) => {
  const enabled = useFlag(flagName);
  const variant = useVariant(flagName);
  const client = useUnleashClient();

  const updateContext = useUnleashContext();
  const { flagsReady, flagsError } = useFlagsStatus();

  const isEnabled = () => {
    return enabled;
  };

  const getVariant = () => {
    return variant;
  };

  const getClient = () => {
    return client;
  };

  const getUnleashContextSetter = () => {
    return updateContext;
  };

  const getFlagsStatus = () => {
    return { flagsReady, flagsError };
  };

  return (
    <>
      {render({
        isEnabled,
        getVariant,
        getClient,
        getUnleashContextSetter,
        getFlagsStatus,
      })}
    </>
  );
};

export default UnleashClassFlagProvider