import * as IconOptions from './Icons';

export type AppIconOptions = keyof typeof IconOptions;
/**
 * Want to add an icon? Follow these steps:
 *
 * 1. Create a new file under the `Icons/` directory with the name of your icon. *I.e. Simulation*
 * 2. Create a __named export__ react component that receives `BaseSVGProps` and returns the icons svg
 * 3. Swap out the svg and path elements with the `BaseSVG` and `BasePath` components. Pass all props from parent to `BaseSVG`
 * 4. Implementation `<AppIcon.Simulation />`
 */
export const AppIcon = IconOptions;
