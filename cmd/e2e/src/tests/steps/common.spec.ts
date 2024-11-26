import { Given, When } from "@cucumber/cucumber";
import { User, IUserResult } from "../../../prisma/seed";

let newUser: Promise<IUserResult>

Given('Create a new user with {string} role', async function (string) {
    const user = new User();
    user.role = string
    newUser = user.create();
});

Given('Create a new user with {string} role with disabled status', async function (string) {
    const user = new User();
    user.role = string
    user.isDisabled = true
    newUser = user.create();
});

Given('User navigates to the login page', async function () {
    await this.fixture.page.goto(`${process.env.BASEURL}/ui/login`);
});

Given('User enters valid username', async function () {
    await this.fixture.page.locator("#username").fill((await newUser).principal_name);
});

Given('User enters valid email', async function () {
    await this.fixture.page.locator("#username").fill((await newUser).email_address);
  });
  
Given('User enters valid password', async function () {
    await this.fixture.page.locator("#password").fill((await newUser).uniquePassword);
});

When('User clicks on the login button', async function () {
    await this.fixture.page.getByRole('button', { name: "LOGIN", exact: true }).click();
});
  