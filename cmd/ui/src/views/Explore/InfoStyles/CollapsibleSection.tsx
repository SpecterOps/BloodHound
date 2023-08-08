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

import makeStyles from '@mui/styles/makeStyles';
import { Theme } from '@mui/material';

const useCollapsibleSectionStyles = makeStyles((theme: Theme) => ({
    accordionRoot: {
        backgroundColor: 'inherit',
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
        color: 'common.black',
        fontSize: '0.75rem',
        '&:hover': {
            backgroundColor: theme.palette.action.hover,
        },
    },
    accordionDetails: {
        padding: theme.spacing(1, 0),
    },
    accordionCount: {
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        fontWeight: 'bold',
        fontSize: '0.9rem',
        backgroundColor: '#d7dee3',
        minWidth: '3rem',
        height: '1.6rem',
        lineHeight: '1.6em',
        paddingX: '0.5rem',
        borderRadius: theme.shape.borderRadius,
    },
    edgeAccordionDetails: {
        padding: theme.spacing(0, 0, 0, 1),
        '& p.MuiTypography-body1': {
            marginTop: '8px',
            fontSize: '0.875rem',
            textDecoration: 'underline',
            padding: theme.spacing(0.5, 0.5),
            borderRadius: theme.shape.borderRadius,
        },
        '& p.MuiTypography-body2': {
            marginTop: '8px',
            fontSize: '0.75rem',
            backgroundColor: '#eee',
            padding: theme.spacing(0.5, 1),
            borderRadius: theme.shape.borderRadius,
        },
        '& a': {
            fontSize: '0.75rem',
            marginTop: '4px',
            whiteSpace: 'nowrap',
        },
        '& pre': {
            fontFamily: '"source-code-pro", "Menlo", "Monaco", "Consolas", "Courier New", "monospace"',
            whiteSpace: 'pre-line',
            fontSize: '0.75rem',
            wordBreak: 'break-all',
            margin: theme.spacing(1, 0.5),
            padding: theme.spacing(0.5),
            backgroundColor: 'rgba(0,0,0,0.75)',
            borderRadius: theme.shape.borderRadius,
            color: '#eee',
        },
        '& ul': {
            marginTop: '8px',
            backgroundColor: '#eee',
            borderRadius: theme.shape.borderRadius,
            listStyle: 'disc',
            padding: theme.spacing(0.5, 3),
        },
        '& li': {
            display: 'list-item',
            padding: 0,
        },
        '& span': {
            fontSize: '0.75rem',
        },
    },
    expandIcon: {
        color: theme.palette.common.black,
    },
    title: {
        marginLeft: theme.spacing(2),
        lineHeight: '3em',
        fontSize: theme.typography.fontSize,
    },
    fieldsContainer: {
        fontSize: '0.75rem',
        '& > :nth-child(odd)': {
            backgroundColor: theme.palette.grey[200],
        },
        borderRadius: theme.shape.borderRadius,
    },
    alertRoot: {
        display: 'flex',
        justifyContent: 'center',
        padding: 0,
        minWidth: '3rem',
        borderRadius: theme.shape.borderRadius,
    },
    alertIcon: {
        padding: '4px',
        margin: 0,
    },
}));

export default useCollapsibleSectionStyles;
