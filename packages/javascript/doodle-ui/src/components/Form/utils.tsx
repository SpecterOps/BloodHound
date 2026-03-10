import * as React from 'react';
import { type FieldPath, type FieldValues } from 'react-hook-form';

type FormFieldContextValue<
    TFieldValues extends FieldValues = FieldValues,
    TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
    name: TName;
};

type FormItemContextValue = {
    id: string;
};

export const FormItemContext = React.createContext<FormItemContextValue>({} as FormItemContextValue);

export const FormFieldContext = React.createContext<FormFieldContextValue>({} as FormFieldContextValue);
