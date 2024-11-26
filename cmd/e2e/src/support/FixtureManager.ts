import { Browser, BrowserContext, Page } from "@playwright/test";
import PageManager from "./pageManager";
import { invokeBrowser } from "./browserManager";

export interface IFixture {
    browser: Browser;
    context: BrowserContext;
    pageManager: PageManager;
    page: Page;
}
// FixtureManager manages the state of the browser and page for each scenario
export default class FixtureManager implements IFixture {
    browser: Browser;
    context: BrowserContext;
    pageManager: PageManager;
    page: Page;

    constructor() {
        this.pageManager = new PageManager();
    }

    get Fixture(): IFixture {
        return {
            browser: this.browser,
            context: this.context,
            pageManager: this.pageManager,
            page: this.pageManager.Page,
        }
    }

    async openBrowser() {
        this.browser = await invokeBrowser();
    }

    async openContext() {
        this.context = await this.browser.newContext();
    }

    async closeContext() {
        this.context.close();
    }

    async closeBrowser() {
        this.browser.close();
    }

    async newPage() {
        await this.pageManager.newPage(this.context);
    }
}
