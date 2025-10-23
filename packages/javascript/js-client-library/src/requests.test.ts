import * as allure from 'allure-js-commons';
import { expect, test } from 'vitest';
import { LoginRequest } from './requests';

test('LoginRequest', async () => {
    await allure.step('Type Assertions Validations', async () => {
        const loginRequest: LoginRequest = { username: 'jdoe', secret: 'secret', login_method: 'secret', otp: 'opts' };
        Object.keys(loginRequest).every(function (key: string) {
            return expect(typeof loginRequest[key as keyof LoginRequest]).toBe('string');
        });
    });
});
