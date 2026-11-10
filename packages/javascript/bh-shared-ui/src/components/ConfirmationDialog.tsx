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
} from 'doodle-ui';
import React, { ReactNode, useCallback, useState } from 'react';

const ConfirmationDialog: React.FC<{
    open: boolean;
    title: string;
    text: string | JSX.Element;
    onCancel: () => void;
    onConfirm: () => void;
    challengeText?: string;
    isLoading?: boolean;
    isConfirmDisabled?: boolean;
    error?: string;
    cancelIcon?: ReactNode;
    confirmIcon?: ReactNode;
    iconPosition?: 'left' | 'right';
    cancelText?: string;
    confirmText?: string;
    confirmDataAnalytics?: string;
}> = ({
    open,
    title,
    text,
    onCancel,
    isLoading,
    isConfirmDisabled,
    error,
    challengeText = '',
    onConfirm,
    cancelIcon,
    confirmIcon,
    iconPosition = 'left',
    cancelText = 'Cancel',
    confirmText = 'Confirm',
    confirmDataAnalytics,
}) => {
    const [challengeTextReply, setChallengeTextReply] = useState<string>('');

    const handleClose = useCallback(() => {
        onCancel();
        setTimeout(() => {
            setChallengeTextReply('');
        }, 1000);
    }, [onCancel]);

    const handleConfirm = useCallback(() => {
        onConfirm();
        setTimeout(() => {
            setChallengeTextReply('');
        }, 1000);
    }, [onConfirm]);

    const renderButtonContent = (buttonText: string, icon?: ReactNode) => (
        <>
            {iconPosition === 'left' && icon}
            {buttonText}
            {iconPosition === 'right' && icon}
        </>
    );

    return (
        <Dialog open={open} data-testid='confirmation-dialog'>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle className='text-lg'>{title}</DialogTitle>
                    <DialogDescription asChild className='text-lg'>
                        {/* Description can contain block elements that cannot be nested inside a paragraph, so using a div asChild */}
                        <div>{text}</div>
                    </DialogDescription>
                    {challengeText && (
                        <DialogDescription asChild className='text-sm'>
                            <div className='pb-1 flex flex-col gap-1'>
                                Please input "{challengeText}" prior to clicking confirm.
                                <Input
                                    placeholder={challengeText}
                                    variant='outlined'
                                    onChange={(e) => setChallengeTextReply(e.target.value)}
                                    value={challengeTextReply}
                                    data-testid='confirmation-dialog_challenge-text'
                                />
                            </div>
                        </DialogDescription>
                    )}
                    <DialogActions>
                        {error && <p className='content-center text-error text-xs mt-[3px]'>{error}</p>}
                        <Button
                            variant='secondary'
                            onClick={handleClose}
                            disabled={isLoading}
                            data-testid='confirmation-dialog_button-no'>
                            {renderButtonContent(cancelText, cancelIcon)}
                        </Button>
                        <Button
                            data-analytics={confirmDataAnalytics}
                            onClick={handleConfirm}
                            disabled={
                                isLoading ||
                                isConfirmDisabled ||
                                challengeText.toLowerCase() !== challengeTextReply.toLowerCase()
                            }
                            data-testid='confirmation-dialog_button-yes'>
                            {renderButtonContent(confirmText, confirmIcon)}
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ConfirmationDialog;
