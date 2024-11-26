module.exports = {
    default: {
        formatOptions: {
            snippetInterface: "async-await"
        },
        paths: [
            "src/tests/features/"
        ],
        dryRun: false,
        require: [
            "src/tests/steps/*.ts",
            "src/support/hooks.ts"
        ],
        requireModule: [
            "ts-node/register"
        ],
        format: [
            "progress-bar",
            "html:test-results/reports/html/cucumber-report.html",
            "json:test-results/reports/json/cucumber-report.json",
            "rerun:@rerun.txt"
        ]
    },
    rerun: {
        formatOptions: {
            snippetInterface: "async-await"
        },
        dryRun: false,
        require: [
            "src/tests/steps/*.ts",
            "src/support/hooks.ts"
        ],
        requireModule: [
            "ts-node/register"
        ],
        format: [
            "progress-bar",
            "html:test-results/reports/html/cucumber-report.html",
            "json:test-results/reports/json/cucumber-report.json",
            "rerun:@rerun.txt"
        ]
    }
}