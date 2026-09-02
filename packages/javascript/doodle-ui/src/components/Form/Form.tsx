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
import * as LabelPrimitive from '@radix-ui/react-label';
import { Slot } from '@radix-ui/react-slot';
import * as React from 'react';
import { Controller, FormProvider, type ControllerProps, type FieldPath, type FieldValues } from 'react-hook-form';
import { Label } from '../Label';
import { cn } from '../utils';
import { useFormField } from './useFormField';
import { FormFieldContext, FormItemContext } from './utils';

const Form = FormProvider;

const FormField = <
    TFieldValues extends FieldValues = FieldValues,
    TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
    ...props
}: ControllerProps<TFieldValues, TName>) => {
    return (
        <FormFieldContext.Provider value={{ name: props.name }}>
            <Controller {...props} />
        </FormFieldContext.Provider>
    );
};

function FormItem({ className, ...props }: React.ComponentProps<'div'>) {
    const id = React.useId();

    return (
        <FormItemContext.Provider value={{ id }}>
            <div data-slot='form-item' className={cn('grid gap-2', className)} {...props} />
        </FormItemContext.Provider>
    );
}

function FormLabel({ className, ...props }: React.ComponentProps<typeof LabelPrimitive.Root>) {
    const { error, formItemId } = useFormField();

    return (
        <Label
            data-slot='form-label'
            data-error={!!error}
            className={cn('data-[error=true]:text-error', className)}
            htmlFor={formItemId}
            {...props}
        />
    );
}

function FormControl({ ...props }: React.ComponentProps<typeof Slot>) {
    const { error, formItemId, formDescriptionId, formMessageId } = useFormField();

    return (
        <Slot
            data-slot='form-control'
            id={formItemId}
            aria-describedby={!error ? `${formDescriptionId}` : `${formDescriptionId} ${formMessageId}`}
            aria-invalid={!!error}
            {...props}
        />
    );
}

function FormDescription({ className, ...props }: React.ComponentProps<'p'>) {
    const { formDescriptionId } = useFormField();

    return <p data-slot='form-description' id={formDescriptionId} className={cn('text-sm', className)} {...props} />;
}

function FormMessage({ className, ...props }: React.ComponentProps<'p'>) {
    const { error, formMessageId } = useFormField();
    const body = error ? String(error?.message ?? '') : props.children;

    if (!body) {
        return null;
    }

    return (
        <p data-slot='form-message' id={formMessageId} className={cn('text-error text-sm', className)} {...props}>
            {body}
        </p>
    );
}

export { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage };
