import { createFileRoute } from '@tanstack/react-router';
import DisabledUser from 'src/views/DisabledUser';

export const Route = createFileRoute('/(auth)/user-disabled')({
    component: DisabledUser,
});
