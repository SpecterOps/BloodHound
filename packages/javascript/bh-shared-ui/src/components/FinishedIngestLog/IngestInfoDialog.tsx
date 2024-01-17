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

import { Button, Dialog, DialogActions, DialogContent } from '@mui/material';

const IngestInfoDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    data: any;
}> = ({ open, onClose, data }) => {
    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            TransitionProps={{
                unmountOnExit: true,
            }}>
            <DialogContent>
                <p>
                    <strong>User:</strong> {data?.user}
                </p>
                <p>
                    <strong>Start Time:</strong> {data?.start_time}
                </p>
                <p>
                    <strong>End Time:</strong> {data?.end_time}
                </p>
                <p>
                    <strong>Status:</strong> {data?.status}
                </p>
                <p>
                    <strong>Status Message:</strong> {data?.status_message}
                </p>
            </DialogContent>
            <DialogActions>
                <Button color='inherit' onClick={onClose}>
                    Close
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default IngestInfoDialog;
