import { BrowserContext, Page } from "@playwright/test";

export default class PageManager {
    currentPage: Page;

    get Page(): Page {
        return this.currentPage;
    }

    async newPage(context: BrowserContext): Promise<Page> {
        const page = await context.newPage();
        this.currentPage = page;
        return page
    }
    async closePage(): Promise<void> {
        await this.currentPage.close();
    }
}
