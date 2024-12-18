import { LaunchOptions, chromium, firefox, webkit } from "@playwright/test";

const options: LaunchOptions = {
    headless: true,
}

export const invokeBrowser = () => {
    // TODO Add option to load browser type configuration at runtime
    const browserType = process.env.DEFAULT_BROWSER;
    switch (browserType) {
        case "chrome":
            return chromium.launch(options);
        case "firefox":
            return firefox.launch(options);
        case "webkit":
            return webkit.launch(options);
        default:
            throw new Error("Please set the proper browser :)")
    }
}