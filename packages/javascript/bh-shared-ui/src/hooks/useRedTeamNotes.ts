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

import {
    RequestOptions,
    RedTeamNote,
    RedTeamNotePayload,
    RedTeamNoteTagCount,
    RedTeamNotesListParams,
} from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils';

export const redTeamNotesKeys = {
    all: ['redTeamNotes'] as const,
    list: (params?: RedTeamNotesListParams) => [...redTeamNotesKeys.all, 'list', { ...(params ?? {}) }] as const,
    detail: (noteId: string | number) => [...redTeamNotesKeys.all, 'detail', noteId] as const,
    tags: () => [...redTeamNotesKeys.all, 'tags'] as const,
};

export const listRedTeamNotes = (
    params: RedTeamNotesListParams,
    options?: RequestOptions
): Promise<{ data: RedTeamNote[]; count: number }> => {
    return apiClient.listRedTeamNotes(params, options).then((response) => ({
        data: response.data.data,
        count: response.data.count,
    }));
};

export const createRedTeamNote = (payload: RedTeamNotePayload, options?: RequestOptions) => {
    return apiClient.createRedTeamNote(payload, options).then((response) => response.data.data);
};

export const updateRedTeamNote = (
    noteId: string | number,
    payload: RedTeamNotePayload,
    options?: RequestOptions
) => {
    return apiClient.updateRedTeamNote(noteId, payload, options).then((response) => response.data.data);
};

export const deleteRedTeamNote = (noteId: string | number, options?: RequestOptions) => {
    return apiClient.deleteRedTeamNote(noteId, options);
};

export function useRedTeamNotes(params: RedTeamNotesListParams) {
    return useQuery({
        queryKey: redTeamNotesKeys.list(params),
        queryFn: ({ signal }) => listRedTeamNotes(params, { signal }),
    });
}

export function useRedTeamNotesByObjectId(objectId: string | undefined) {
    return useQuery({
        queryKey: redTeamNotesKeys.list({ object_id: objectId }),
        queryFn: ({ signal }) => listRedTeamNotes({ object_id: objectId, limit: 100 }, { signal }),
        enabled: !!objectId,
    });
}

export function useRedTeamNotesByEdgeKind(edgeKind: string | undefined) {
    return useQuery({
        queryKey: redTeamNotesKeys.list({ edge_kind: edgeKind }),
        queryFn: ({ signal }) => listRedTeamNotes({ edge_kind: edgeKind, limit: 100 }, { signal }),
        enabled: !!edgeKind,
    });
}

export function useCreateRedTeamNote() {
    const queryClient = useQueryClient();

    return useMutation(createRedTeamNote, {
        onSuccess: () => {
            queryClient.invalidateQueries(redTeamNotesKeys.all);
        },
    });
}

export function useUpdateRedTeamNote() {
    const queryClient = useQueryClient();

    return useMutation(({ noteId, payload }: { noteId: string | number; payload: RedTeamNotePayload }) =>
        updateRedTeamNote(noteId, payload)
    , {
        onSuccess: () => {
            queryClient.invalidateQueries(redTeamNotesKeys.all);
        },
    });
}

export function useDeleteRedTeamNote() {
    const queryClient = useQueryClient();

    return useMutation(deleteRedTeamNote, {
        onSuccess: () => {
            queryClient.invalidateQueries(redTeamNotesKeys.all);
        },
    });
}

export const listRedTeamNoteTags = (options?: RequestOptions): Promise<RedTeamNoteTagCount[]> => {
    return apiClient.listRedTeamNoteTags(options).then((response) => response.data.data);
};

export const uploadRedTeamNoteAttachment = (file: File, options?: RequestOptions) => {
    return apiClient.uploadRedTeamNoteAttachment(file, options).then((response) => response.data.data);
};

export function useRedTeamNoteTags() {
    return useQuery({
        queryKey: redTeamNotesKeys.tags(),
        queryFn: ({ signal }) => listRedTeamNoteTags({ signal }),
    });
}
