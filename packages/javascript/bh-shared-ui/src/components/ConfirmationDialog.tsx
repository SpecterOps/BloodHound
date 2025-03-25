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

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
} from '@bloodhoundenterprise/doodleui';
import { FormHelperText } from '@mui/material';
import React, { useCallback, useState } from 'react';

const ConfirmationDialog: React.FC<{
    open: boolean;
    title: string;
    text: string | JSX.Element;
    onClose: (response: boolean) => void;
    challengeTxt?: string;
    isLoading?: boolean;
    error?: string;
}> = ({ open, title, text, onClose, isLoading, error, challengeTxt }) => {
    const [challengeTxtReply, setChallengeTxtReply] = useState<string>('');

    const handleClose = useCallback(
        (response: boolean) => () => {
            onClose(response);
            setTimeout(() => {
                setChallengeTxtReply('');
            }, 1000);
        },
        [onClose]
    );

    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle className='text-lg'>{title}</DialogTitle>
                    <DialogDescription className='text-lg'>{text}</DialogDescription>
                    {challengeTxt && (
                        <>
                            <DialogDescription className='text-sm'>
                                Please input "{challengeTxt}" prior to clicking confirm.
                                <Input
                                    placeholder={challengeTxt.toLowerCase()}
                                    className='border-t-0 border-l-0 border-r-0 rounded-none border-black dark:border-white bg-transparent dark:bg-transparent placeholder-neutral-dark-10 dark:placeholder-neutral-light-10 focus-visible:ring-0 focus-visible:ring-offset-0 pl-2'
                                    onChange={(e) => setChallengeTxtReply(e.target.value)}
                                    value={challengeTxtReply}
                                    data-testid='confirmation-dialog_challenge-text'
                                />
                            </DialogDescription>
                        </>
                    )}
                    <DialogActions>
                        {error && (
                            <FormHelperText error className='content-center'>
                                {error}
                            </FormHelperText>
                        )}
                        <Button
                            variant='tertiary'
                            onClick={handleClose(false)}
                            disabled={isLoading}
                            data-testid='confirmation-dialog_button-no'>
                            Cancel
                        </Button>
                        <Button
                            onClick={handleClose(true)}
                            disabled={
                                isLoading ||
                                (!!challengeTxt && challengeTxt.toLowerCase() !== challengeTxtReply.toLowerCase())
                            }
                            data-testid='confirmation-dialog_button-yes'>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ConfirmationDialog;
