import * as allure from 'allure-js-commons';
import { expect, test } from 'vitest';
import { LoginRequest } from './requests';

test('LoginRequest', async () => {
    await allure.step('Type Assertions Validations', async () => {
        const loginRequest: LoginRequest = { username: 'jdoe', secret: 'secret', login_method: 'secret', otp: 'opts' };
        for (const value of Object.values(loginRequest)) {
            expect(typeof value).toBe('string');
        }
    });
});
