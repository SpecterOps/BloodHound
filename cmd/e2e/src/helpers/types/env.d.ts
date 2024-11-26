export { };

declare global {
    namespace NodeJS {
        interface ProcessEnv {
            DEFAULT_BROWSER: "chrome" | "firefox" | "webkit",
            ENV: "staging" | "production" | "dev",
            BASEURL: string
        }
    }
}