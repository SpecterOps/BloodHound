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

import React from 'react';
import { Typography } from '@mui/material';

const escapeSpecialCharacters = (text: string) => text.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');

const HighlightedText: React.FC<{ text: string; search: string }> = ({ text, search }) => {
    const escapedSearch = escapeSpecialCharacters(search);
    const regex = new RegExp(`(.*?)(${escapedSearch})(.*)`, 'mi');
    const groups = text.match(regex);
    if (groups === null || groups.length === 1) return <>{text}</>;

    const parts = groups.slice(1);

    return (
        <Typography variant='body2' component={'span'}>
            {parts.map((part, i) => {
                if (part.toLowerCase() === search.toLowerCase())
                    return (
                        <Typography variant='body2' component={'span'} style={{ fontWeight: 'bold' }} key={i}>
                            {part}
                        </Typography>
                    );
                return <React.Fragment key={i}>{part}</React.Fragment>;
            })}
        </Typography>
    );
};

export default HighlightedText;
