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

import { Theme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';

const useCollapsibleSectionStyles = makeStyles((theme: Theme) => ({
    accordionRoot: {
        backgroundColor: 'inherit',
        backgroundImage: 'unset',
        margin: 0,
        '&.Mui-disabled': {
            backgroundColor: 'inherit',
        },
        '&.Mui-expanded': {
            margin: 0,
        },
    },
    accordionSummary: {
        padding: theme.spacing(0, 2),
        margin: theme.spacing(0, -2),
        fontSize: '0.75rem',
        '&:hover': {
            backgroundColor: theme.palette.action.hover,
        },
    },
    accordionDetails: {
        padding: theme.spacing(1, 0),
    },
    edgeAccordionDetails: {
        padding: theme.spacing(0, 0, 0, 1),
        '& p.typography-body1': {
            marginTop: '8px',
            fontSize: '0.875rem',
            textDecoration: 'underline',
            padding: theme.spacing(0.5, 0.5),
            borderRadius: theme.shape.borderRadius,
        },
        '& p.typography-body2, & p.edge-accordion-body2': {
            marginTop: '8px',
            fontSize: '0.75rem',
            lineHeight: '1.43',
            backgroundColor: theme.palette.neutral.tertiary,
            padding: theme.spacing(0.5, 1),
            borderRadius: theme.shape.borderRadius,
        },
        '& a': {
            fontSize: '0.75rem',
            marginTop: '4px',
            whiteSpace: 'normal',
            overflowWrap: 'anywhere',
            wordBreak: 'break-word',
            textDecoration: 'underline',
            color: theme.palette.color.links,
            cursor: 'pointer',
        },
        '& pre': {
            fontFamily: '"source-code-pro", "Menlo", "Monaco", "Consolas", "Courier New", "monospace"',
            whiteSpace: 'pre-line',
            fontSize: '0.75rem',
            lineHeight: '1.5',
            wordBreak: 'break-all',
            margin: theme.spacing(1, 0),
            padding: theme.spacing(0.5, 1),
            backgroundColor: theme.palette.neutral.quinary,
            borderRadius: theme.shape.borderRadius,
            color: theme.palette.color.primary,
        },
        '& ul': {
            marginTop: '8px',
            backgroundColor: theme.palette.neutral.tertiary,
            borderRadius: theme.shape.borderRadius,
            listStyle: 'disc',
            padding: theme.spacing(0.5, 3),
        },
        '& span': {
            fontSize: '0.75rem',
        },
    },
    expandIcon: {
        color: theme.palette.color.primary,
    },
}));

export default useCollapsibleSectionStyles;
