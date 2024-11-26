import { setWorldConstructor, World } from "@cucumber/cucumber";
import { IFixture } from "../../support/FixtureManager";

// Extend World base class properties and methods with PlayWrightWorld Class
export default class PlaywrightWorld extends World {
    fixture: IFixture;
}

setWorldConstructor(PlaywrightWorld);
