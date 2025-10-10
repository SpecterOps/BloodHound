// Copyright 2025 Specter Ops, Inc.
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
import { Card, CardContent, CardFooter, CardTitle } from '@bloodhoundenterprise/doodleui';
import { AppIcon } from '../../../components';
import { useHistoryTableContext } from './HistoryTableContext';

const HistoryNote = () => {
    const { currentNote } = useHistoryTableContext();

    return (
        <div>
            <Card className='flex justify-center mb-4 p-4 h-[56px] w-[400px] min-w-[300px] max-w-[32rem]'>
                <CardTitle className='flex  items-center gap-2'>
                    <AppIcon.LinedPaper size={24} />
                    Note
                </CardTitle>
            </Card>

            {currentNote && (
                <Card className='p-4 '>
                    <CardContent>
                        <p className='text-xl'>{currentNote.note}</p>
                    </CardContent>
                    <CardFooter>
                        <p>
                            By {currentNote.createdBy} on {currentNote.timestamp}
                        </p>
                    </CardFooter>
                </Card>
            )}
        </div>
    );
};

export default HistoryNote;
