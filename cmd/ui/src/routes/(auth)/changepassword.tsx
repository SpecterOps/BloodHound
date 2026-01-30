import { createFileRoute } from '@tanstack/react-router';
import Login from 'src/views/Login';

export const Route = createFileRoute('/(auth)/changepassword')({
    component: Login,
});
