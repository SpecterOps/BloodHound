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

import { PureComponent } from 'react';
import PropTypes from 'prop-types';
import { List } from 'immutable';
import ImPropTypes from 'react-immutable-proptypes';
import { CommunityIcon } from './CommunityIcon';
import { EnterpriseIcon } from './EnterpriseIcon';
import toString from 'lodash/toString';

export const OperationsEditionPlugin = function () {
    return {
        wrapComponents: {
            OperationSummary: (Original: any, system: any) => (props: any) => {
                // The component only has access to the tag that is currently being rendered and not the entire array.
                // This looks up the array by the top-level system attribute so it can be passed into the component at render time.
                const [, path, action] = props.specPath.toJS();
                const tags = system.spec().toJS().json.paths[path][action].tags;
                const isCommunity = tags.includes('Community');
                const isEnterprise = tags.includes('Enterprise');

                return (
                    <div>
                        <OperationSummaryWithEdition {...props} isCommunity={isCommunity} isEnterprise={isEnterprise} />
                    </div>
                );
            },
        },
    };
};

// This component is a copy of the default of OperationSummary component from the https://github.com/swagger-api/swagger-ui repository.
// It adds an additional element displaying the different editions that the operation is available in based on the presence
// of the "Community" and "Enterprise" tags on the endpoint definition.
// It also removed the CopyToClipboardBtn as the show/hide aspect makes it difficult to display the hovertext on the edition badges.
export class OperationSummaryWithEdition extends PureComponent<{
    isShown: any;
    toggleShown: any;
    getComponent: any;
    authActions: any;
    authSelectors: any;
    operationProps: any;
    specPath: any;
    isCommunity: boolean;
    isEnterprise: boolean;
}> {
    static propTypes = {
        specPath: ImPropTypes.list.isRequired,
        operationProps: PropTypes.any.isRequired,
        isShown: PropTypes.bool.isRequired,
        toggleShown: PropTypes.func.isRequired,
        getComponent: PropTypes.func.isRequired,
        authActions: PropTypes.object,
        authSelectors: PropTypes.object,
    };

    static defaultProps = {
        operationProps: null,
        specPath: List(),
        summary: '',
    };

    render() {
        const {
            isShown,
            toggleShown,
            getComponent,
            authActions,
            authSelectors,
            operationProps,
            specPath,
            isCommunity,
            isEnterprise,
        } = this.props;

        const {
            summary,
            isAuthorized,
            method,
            op,
            showSummary,
            path,
            operationId,
            originalOperationId,
            displayOperationId,
        } = operationProps.toJS();

        const { summary: resolvedSummary } = op;

        const security = operationProps.get('security');

        const AuthorizeOperationBtn = getComponent('authorizeOperationBtn');
        const OperationSummaryMethod = getComponent('OperationSummaryMethod');
        const OperationSummaryPath = getComponent('OperationSummaryPath');
        const JumpToPath = getComponent('JumpToPath', true);

        const hasSecurity = security && !!security.count();
        const securityIsOptional = hasSecurity && security.size === 1 && security.first().isEmpty();
        const allowAnonymous = !hasSecurity || securityIsOptional;
        return (
            <div className={`opblock-summary opblock-summary-${method}`}>
                <button
                    aria-label={`${method} ${path.replace(/\//g, '\u200b/')}`}
                    aria-expanded={isShown}
                    className='opblock-summary-control'
                    onClick={toggleShown}>
                    <OperationSummaryMethod method={method} />
                    <OperationSummaryPath
                        getComponent={getComponent}
                        operationProps={operationProps}
                        specPath={specPath}
                    />

                    {!showSummary ? null : (
                        <div className='opblock-summary-description'>{toString(resolvedSummary || summary)}</div>
                    )}

                    <CommunityIcon
                        style={{ marginRight: '10px' }}
                        fill={isCommunity ? '#EE290D' : 'grey'}
                        title={
                            isCommunity
                                ? 'Available in BloodHound Community Edition'
                                : 'Not available in BloodHound Community Edition'
                        }
                        width='50px'
                        height='33px'
                    />
                    <EnterpriseIcon
                        style={{ marginRight: '15px' }}
                        fill={isEnterprise ? '#34318F' : 'grey'}
                        title={
                            isEnterprise
                                ? 'Available in BloodHound Enterprise'
                                : 'Not available in BloodHound Enterprise'
                        }
                        width='47px'
                        height='30px'
                    />

                    {displayOperationId && (originalOperationId || operationId) ? (
                        <span className='opblock-summary-operation-id'>{originalOperationId || operationId}</span>
                    ) : null}

                    <svg className='arrow' width='20' height='20' aria-hidden='true' focusable='false'>
                        <use
                            href={isShown ? '#large-arrow-up' : '#large-arrow-down'}
                            xlinkHref={isShown ? '#large-arrow-up' : '#large-arrow-down'}
                        />
                    </svg>
                </button>

                {allowAnonymous ? null : (
                    <AuthorizeOperationBtn
                        isAuthorized={isAuthorized}
                        onClick={() => {
                            const applicableDefinitions = authSelectors.definitionsForRequirements(security);
                            authActions.showDefinitions(applicableDefinitions);
                        }}
                    />
                )}

                <JumpToPath path={specPath} />
            </div>
        );
    }
}
