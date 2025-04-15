// // Copyright 2025 Specter Ops, Inc.
// //
// // Licensed under the Apache License, Version 2.0
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.
// //
// // SPDX-License-Identifier: Apache-2.0

// import userEvent from '@testing-library/user-event';
// import { SearchValue } from '../../store';
// import { act, render, screen } from '../../test-utils';
// import AssetGroupSelectorObjectSelect from './AssetGroupSelectorObjectSelect';

// describe('AssetGroupSelectorObjectSelect', () => {
//     const user = userEvent.setup();
//     const selectedNodes: (SearchValue & { memberCount?: number })[] = [
//         {
//             name: 'bruce@gotham.local',
//             objectid: '1',
//             type: 'User',
//             memberCount: 777,
//         },
//     ];
//     const onDeleteNode = vi.fn();
//     const onSelectNode = vi.fn();
//     beforeEach(async () => {
//         await act(async () => {
//             render(
//                 <AssetGroupSelectorObjectSelect
//                     selectedNodes={selectedNodes}
//                     onDeleteNode={onDeleteNode}
//                     onSelectNode={onSelectNode}
//                 />
//             );
//         });
//     });

//     it('should render', async () => {
//         expect(
//             await screen.findByText(
//                 /use the input field to add objects and the edit button to remove objects from the list/i
//             )
//         ).toBeInTheDocument();
//         expect(await screen.findByTestId('explore_search_input-search')).toBeInTheDocument();
//         expect(await screen.findByTestId('selector-object-search_edit-button')).toBeInTheDocument();
//         expect(await screen.findByText(`${selectedNodes[0].name}`)).toBeInTheDocument();
//         expect(await screen.findByText(/777 members/i)).toBeInTheDocument();
//     });

//     it('invokes onDeleteNode when clicked', async () => {
//         const editBtn = await screen.findByTestId('selector-object-search_edit-button');
//         await user.click(editBtn);

//         const deleteBtn = await screen.findByText('trash-can');

//         await user.click(deleteBtn);

//         expect(onDeleteNode).toBeCalledTimes(1);
//         expect(onDeleteNode).toBeCalledWith(selectedNodes[0].objectid);
//     });
// });
