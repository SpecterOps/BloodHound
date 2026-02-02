import { createLazyFileRoute } from '@tanstack/react-router';
import ExpiredPassword from 'src/views/ExpiredPassword';

export const Route = createLazyFileRoute('/(auth)/expired-password')({
    component: ExpiredPassword,
});
