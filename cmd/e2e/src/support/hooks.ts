import { Before, After, BeforeAll, AfterAll, Status } from "@cucumber/cucumber";
import FixtureManager from "./FixtureManager.js";
import PlaywrightWorld from "../tests/worlds/playwrightWorld.js";
import { loadEnvs } from "../helpers/env/env.js";
import { dbOPS } from "../../prisma/cleanup.js";

let fx: FixtureManager

BeforeAll(async function () {
    // load environment variables
    loadEnvs();

    fx = new FixtureManager();
    await fx.openBrowser();
})

Before(async function (this: PlaywrightWorld) {
    // create new instance of fixture for each scenario
    await fx.openContext();
    await fx.newPage();
    this.fixture = fx.Fixture;
});

After(async function ({ result, pickle }) {
    // capture screenshot for failed step
    if (result?.status == Status.FAILED) {
        const img = await this.fixture.page.screenshot({ path: `./test-results/screenshots/+${pickle.name}`, type: "png" })
        await this.attach(img, "image/png");
    }
    
    // delete test users in dev environment
    const db = new dbOPS();
    db.deleteUsers();

});

AfterAll(async function () {
    await fx.closeBrowser();
})
