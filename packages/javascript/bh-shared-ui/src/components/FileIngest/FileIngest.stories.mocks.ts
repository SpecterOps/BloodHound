// Copyright 2026 Specter Ops, Inc.
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
import type { FileIngestCompletedTask, FileIngestJob } from 'js-client-library';
import { rest } from 'msw';
import { ZERO_VALUE_API_DATE } from '../../constants';
import { PERMISSIONS, Permission } from '../../utils';
import { FileUploadJobStatus } from '../LegacyFileIngestTable/types';

export type IngestScenario = {
    feature?: 'enabled' | 'disabled' | 'loading' | 'error' | 'missing';
    permission?: 'manage' | 'read' | 'none';
    history?: 'populated' | 'empty' | 'loading' | 'error';
    upload?: 'success' | 'file-error';
};

const timestamps = {
    created_at: '2026-09-21T16:00:00Z',
    updated_at: '2026-09-21T16:03:00Z',
    deleted_at: { Time: ZERO_VALUE_API_DATE, Valid: false },
};

export const ingestUsers = [
    { id: 'story-collector', email_address: 'collector@example.com' },
    { id: 'story-analyst', email_address: 'analyst@example.com' },
];

const ingestStatuses = [
    FileUploadJobStatus.PARTIALLY_COMPLETE,
    FileUploadJobStatus.COMPLETE,
    FileUploadJobStatus.FAILED,
    FileUploadJobStatus.INGESTING,
    FileUploadJobStatus.ANALYZING,
    FileUploadJobStatus.RUNNING,
    FileUploadJobStatus.READY,
    FileUploadJobStatus.CANCELED,
    FileUploadJobStatus.TIMED_OUT,
    FileUploadJobStatus.INVALID,
];
const activeStatuses = [
    FileUploadJobStatus.READY,
    FileUploadJobStatus.RUNNING,
    FileUploadJobStatus.INGESTING,
    FileUploadJobStatus.ANALYZING,
];

// Two pages, every status, two users, and fixed dates make filtering and pagination reproducible.
export const createIngestJobs = (): FileIngestJob[] =>
    Array.from({ length: 24 }, (_, index) => {
        const status = ingestStatuses[index % ingestStatuses.length];
        const user = ingestUsers[index % 2];
        const start = `2026-09-${String(24 - Math.floor(index / 4)).padStart(2, '0')}T16:00:00Z`;
        const active = activeStatuses.includes(status);
        return {
            ...timestamps,
            id: 124 - index,
            user_id: user.id,
            user_email_address: user.email_address,
            status,
            status_message:
                status === FileUploadJobStatus.PARTIALLY_COMPLETE
                    ? 'Completed with warnings'
                    : status === FileUploadJobStatus.FAILED
                      ? 'Invalid collector data'
                      : 'Collector upload',
            start_time: start,
            end_time: active ? ZERO_VALUE_API_DATE : start.replace('16:00', '16:03'),
            last_ingest: start,
            total_files: 3,
            failed_files: status === FileUploadJobStatus.FAILED ? 1 : 0,
        };
    });

const createTasks = (job: FileIngestJob): FileIngestCompletedTask[] =>
    ['users.json', 'groups.json', 'computers.json'].map((file_name, index) => ({
        ...timestamps,
        id: job.id * 10 + index,
        file_name,
        parent_file_name: 'sharphound.zip',
        errors:
            job.status === FileUploadJobStatus.FAILED && index === 2
                ? ['Invalid collector schema: missing metadata.']
                : [],
        warnings:
            job.status === FileUploadJobStatus.PARTIALLY_COMPLETE && index === 1
                ? ['Some objects could not be resolved.']
                : [],
    }));

export const createFileIngestHandlers = ({
    feature = 'enabled',
    permission = 'manage',
    history = 'populated',
    upload = 'success',
}: IngestScenario = {}) => {
    const permissions =
        permission === 'none'
            ? []
            : [
                  PERMISSIONS[
                      permission === 'manage' ? Permission.GRAPH_DB_INGEST_MANAGE : Permission.GRAPH_DB_INGEST_READ
                  ],
              ];
    let jobs = history === 'empty' ? [] : createIngestJobs();
    const uploadedTasks = new Map<number, FileIngestCompletedTask[]>();
    const failedOnce = new Set<string>();
    let nextId = 125;

    return [
        rest.get('/api/v2/self', (_request, response, context) =>
            response(
                context.json({
                    data: {
                        ...ingestUsers[0],
                        principal_name: 'story-collector',
                        roles: [{ name: permission, permissions }],
                    },
                })
            )
        ),
        rest.get('/api/v2/features', (_request, response, context) => {
            if (feature === 'loading') return response(context.delay('infinite'));
            if (feature === 'error') return response(context.status(500));
            return response(
                context.json({
                    data:
                        feature === 'missing'
                            ? []
                            : [
                                  {
                                      id: 1,
                                      key: 'open_graph_phase_2',
                                      name: 'OpenGraph phase 2',
                                      description: 'File ingest details and filters',
                                      enabled: feature === 'enabled',
                                      user_updatable: true,
                                  },
                              ],
                })
            );
        }),
        rest.get('/api/v2/bloodhound-users-minimal', (_request, response, context) =>
            response(context.json({ data: { users: ingestUsers } }))
        ),
        rest.get('/api/v2/file-upload', (request, response, context) => {
            if (history === 'loading') return response(context.delay('infinite'));
            if (history === 'error') return response(context.status(500));
            const params = request.url.searchParams;
            const value = (key: string) => params.get(key)?.replace(/^(eq|gte|lte):/, '');
            const filtered = jobs
                .filter(
                    (job) =>
                        (!value('status') || job.status === Number(value('status'))) &&
                        (!value('user_id') || job.user_id === value('user_id')) &&
                        (!value('start_time') || Date.parse(job.start_time) >= Date.parse(value('start_time')!)) &&
                        (!value('end_time') || Date.parse(job.end_time) <= Date.parse(value('end_time')!))
                )
                .sort((a, b) => b.id - a.id);
            const skip = Number(params.get('skip') ?? 0);
            const limit = Number(params.get('limit') ?? 10);
            return response(
                context.json({ data: filtered.slice(skip, skip + limit), count: filtered.length, limit, skip })
            );
        }),
        rest.get('/api/v2/file-upload/:id/completed-tasks', (request, response, context) => {
            const id = Number(request.params.id);
            const job = jobs.find((job) => job.id === id);
            return response(context.json({ data: uploadedTasks.get(id) ?? (job ? createTasks(job) : []) }));
        }),
        rest.get('/api/v2/file-upload/accepted-types', (_request, response, context) =>
            response(context.json({ data: ['application/json', 'application/zip'] }))
        ),
        rest.post('/api/v2/file-upload/start', (_request, response, context) => {
            const job: FileIngestJob = {
                ...createIngestJobs()[0],
                id: nextId++,
                status: FileUploadJobStatus.READY,
                status_message: 'Ready',
                total_files: 0,
                failed_files: 0,
                end_time: ZERO_VALUE_API_DATE,
            };
            jobs = [job, ...jobs];
            uploadedTasks.set(job.id, []);
            return response(context.status(201), context.json({ data: job }));
        }),
        rest.post('/api/v2/file-upload/:id', (request, response, context) => {
            const id = Number(request.params.id);
            const name = request.headers.get('X-File-Upload-Name') ?? 'upload.json';
            const key = `${id}/${name}`;
            // Fail once so the real retry action can recover without changing handlers.
            if (upload === 'file-error' && !failedOnce.has(key)) {
                failedOnce.add(key);
                return response(
                    context.status(400),
                    context.json({
                        errors: [{ message: 'The collector file could not be uploaded. Retry this file.' }],
                    })
                );
            }
            const tasks = uploadedTasks.get(id) ?? [];
            tasks.push({
                ...timestamps,
                id: id * 10 + tasks.length,
                file_name: name,
                parent_file_name: '',
                errors: [],
                warnings: [],
            });
            uploadedTasks.set(id, tasks);
            const job = jobs.find((job) => job.id === id);
            if (job) job.total_files = tasks.length;
            return response(context.delay(250), context.status(202), context.json({ data: null }));
        }),
        rest.post('/api/v2/file-upload/:id/end', (request, response, context) => {
            const job = jobs.find((job) => job.id === Number(request.params.id));
            if (job) {
                job.status = FileUploadJobStatus.COMPLETE;
                job.status_message = 'Complete';
                job.end_time = job.start_time.replace('16:00', '16:03');
                job.total_files = uploadedTasks.get(job.id)?.length ?? 0;
            }
            return response(context.json({ data: null }));
        }),
    ];
};
