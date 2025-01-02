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

import { Dialog, DialogContent, DialogDescription, DialogPortal, DialogTitle } from '@bloodhoundenterprise/doodleui';

type NoDataDialogProps = {
    fileIngestLink: JSX.Element;
    gettingStartedLink: JSX.Element;
    open: boolean;
};

export const NoDataDialog: React.FC<NoDataDialogProps> = ({ fileIngestLink, gettingStartedLink, open }) => {
    return (
        <Dialog
            open={open}
            onOpenChange={() => {
                // unblocks the body from being clickable so the user can go to another tab
                document.body.style.pointerEvents = '';
            }}>
            <DialogPortal>
                <DialogContent className='focus:outline-none' DialogOverlayProps={{ className: 'top-12' }}>
                    <DialogTitle>No Data Available</DialogTitle>
                    <DialogDescription>
                        To explore your environment, {fileIngestLink}, on the file ingest page. Need help? Check out the{' '}
                        {gettingStartedLink} for instructions.
                    </DialogDescription>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};
