import fs from 'fs';
import path from 'path';

const urlVersion = '__URL_VERSION__';
const version = '__VERSION__';
const templatePrefix = '<img src="https://img.shields.io/badge/version-';

const badgeTemplate = `${templatePrefix}${urlVersion}-teal" alt="version ${version}"/>`;

const main = () => {
    const args = process.argv.slice(2);
    const versionString = args[0];
    const readmePath = path.join(args[1], 'README.md');

    let newBadgeTag = badgeTemplate.replace(version, versionString);
    newBadgeTag = newBadgeTag.replace(urlVersion, versionString.replace('-', '--'));

    const readmeLines = fs.readFileSync(readmePath, 'utf-8').split('\n');

    let versionBadgeLineIndex = -1;
    readmeLines.forEach((line, index) => {
        if (line.includes(templatePrefix)) versionBadgeLineIndex = index;
        return;
    });

    if (versionBadgeLineIndex < 0) {
        console.warn(`Did not find the img tag for replacing the version in.`);
        return;
    }

    if (readmeLines[versionBadgeLineIndex] !== newBadgeTag) {
        console.log(
            `
            
Replacing the version badge in the README with version ${versionString}. Please make sure to commit the changes!
            
            `
        );
        readmeLines[versionBadgeLineIndex] = newBadgeTag;
        fs.writeFileSync(readmePath, readmeLines.join('\n'));
    }
};

main();
