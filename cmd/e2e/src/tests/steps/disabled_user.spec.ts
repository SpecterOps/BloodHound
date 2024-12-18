import { Given, When, Then, setDefaultTimeout } from "@cucumber/cucumber";
import { expect } from "@playwright/test";

Given('User visits the login page', async function () {
  await this.fixture.page.goto(`${process.env.BASEURL}/ui/login`);
});

Then('login page displays {string}', async function (string) {
  await expect(this.fixture.page.getByText(string)).toBeVisible();
});