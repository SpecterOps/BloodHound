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

import {
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Input,
    Label,
    Skeleton,
    Switch,
} from '@bloodhoundenterprise/doodleui';
import { createBrowserHistory } from 'history';
import { FC } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { apiClient } from '../../../utils';

type TagFormInputs = {
    name: string;
    description: string;
    certificationRequired: boolean;
};

export const TagForm: FC = () => {
    const { tagId = '' } = useParams();
    const history = createBrowserHistory();
    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<TagFormInputs>();

    const navigate = useNavigate();

    const onSubmit: SubmitHandler<TagFormInputs> = (data) => {
        console.log(data);
        console.log(errors);

        navigate(`/tier-management/details/tag/${tagId}`);
    };

    const tagQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTag(tagId);
            return response.data.data;
        },
        enabled: tagId !== '',
    });

    if (tagQuery.isLoading) return <Skeleton />;
    if (tagQuery.isError) throw new Error();

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Card className='min-w-96 w-[672px] p-3 mt-6'>
                <CardHeader>
                    <CardTitle>Tier Details</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className='flex justify-between'>
                        <span>Tier Information</span>
                    </div>
                    <div onSubmit={handleSubmit(onSubmit)} className='flex flex-col gap-6 mt-6'>
                        <div>
                            <Label htmlFor='name'>Name</Label>
                            <Input
                                id='name'
                                type='text'
                                {...register('name', { required: true, value: tagQuery.data?.tag.name })}
                                className={
                                    'rounded-none text-base bg-transparent dark:bg-transparent border-t-0 border-x-0 border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b-[1px] focus-visible:outline-none focus:border-t-0 focus:border-x-0 focus-visible:ring-offset-0 focus-visible:ring-transparent focus-visible:border-secondary focus-visible:border-b-2 focus:border-secondary focus:border-b-2 dark:focus-visible:outline-none dark:focus:border-t-0 dark:focus:border-x-0 dark:focus-visible:ring-offset-0 dark:focus-visible:ring-transparent dark:focus-visible:border-secondary-variant-2 dark:focus-visible:border-b-2 dark:focus:border-secondary-variant-2 dark:focus:border-b-2 hover:border-b-2'
                                }
                            />
                            {errors.name && (
                                <p className='text-sm text-rose-700'>Please provide a name value for your tag</p>
                            )}
                        </div>
                        <div>
                            <Label htmlFor='description'>Description</Label>
                            <textarea
                                id='description'
                                {...register('description', { value: tagQuery.data?.tag.description })}
                                placeholder='Description Input'
                                rows={3}
                                className={
                                    'rounded-md dark:bg-neutral-dark-5 pl-2 w-full mt-2 focus-visible:outline-none focus:ring-secondary focus-visible:ring-secondary focus:outline-secondary focus-visible:outline-secondary dark:focus:ring-secondary-variant-2 dark:focus-visible:ring-secondary-variant-2 dark:focus:outline-secondary-variant-2 dark:focus-visible:outline-secondary-variant-2'
                                }
                            />
                        </div>
                        <div className='hidden'>
                            <Label htmlFor='certificationRequired'>Required Certificaiton</Label>
                            <Switch
                                id='certificationRequired'
                                {...register('certificationRequired', {
                                    value: tagQuery.data?.tag.requireCertify || true,
                                })}
                                label='Enable this to mandate certification for all objects within this tier'
                                defaultChecked></Switch>
                        </div>
                    </div>
                </CardContent>
            </Card>
            <div className='flex justify-end gap-6 mt-6 w-[672px]'>
                <Button variant={'secondary'} onClick={history.back}>
                    Cancel
                </Button>
                <Button
                    variant={'primary'}
                    onClick={() => {
                        navigate(`/tier-management/edit/tag/${tagId}/selector`);
                    }}>
                    Define Selector
                </Button>
            </div>
        </form>
    );
};
