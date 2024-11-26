import { LaunchOptions, chromium, firefox, webkit } from "@playwright/test";

const options: LaunchOptions = {
    headless: true,
}

export const invokeBrowser = () => {
    // Browser configuration supports two inputs from runtime --BROWSER flag and cross-env BROWSER env variable 
    const browserType = process.env.npm_config_browser || process.env.DEFAULT_BROWSER;
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