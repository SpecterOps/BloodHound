import { Given, When, Then, setDefaultTimeout } from "@cucumber/cucumber";
import { expect } from "@playwright/test";

Then('User is redirect to {string} page', async function (string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(string +'$'));
});

Given('User enters invalid password', async function () {
  await this.fixture.page.locator("#password").fill("test1234");
});

When('User visits {string} page', async function (string) {
  await this.fixture.page.goto(`${process.env.BASEURL}/ui/${string}`);
});

When('User reloads the {string} page', async function (string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(string+'$'));
  await this.fixture.page.reload({ waitUntil: "domcontentloaded" });
});

Then('User should be logged', async function () {
  await expect(this.fixture.page).toHaveURL(new RegExp('explore$'));
});

Then('User is redirect back to {string} page', async function (string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(string + '$'));
});

Then('Page Displays Error Message', async function () {
  await expect(this.fixture.page.locator('#notistack-snackbar')).toBeVisible();
});
