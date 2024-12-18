import { Given, When, Then, setDefaultTimeout } from "@cucumber/cucumber";
import { expect } from "@playwright/test";

Then('User is redirect to {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text +'$'));
});

Given('User enters invalid password', async function () {
  await this.fixture.page.locator("#password").fill("test1234");
});

When('User visits {string} page', async function (text: string) {
  await this.fixture.page.goto(`${process.env.BASEURL}/ui/${text}`);
});

When('User reloads the {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text+'$'));
  await this.fixture.page.reload({ waitUntil: "domcontentloaded" });
});

Then('User should be logged', async function () {
  await expect(this.fixture.page).toHaveURL(new RegExp('explore$'));
});

Then('User is redirect back to {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text + '$'));
});

Then('Page Displays Error Message', async function () {
  await expect(this.fixture.page.locator('#notistack-snackbar')).toBeVisible();
});
