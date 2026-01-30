import { createFileRoute } from '@tanstack/react-router';
import ExpiredPassword from 'src/views/ExpiredPassword';

export const Route = createFileRoute('/(auth)/expired-password')({
    component: ExpiredPassword,
});
