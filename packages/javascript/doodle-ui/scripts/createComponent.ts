import fs from 'fs';
import path from 'path';

const replaceInFile = (filePath: string, searchString: string, replaceString: string) => {
    const fileContent = fs.readFileSync(filePath, 'utf8');
    const replacedContent = fileContent.replace(new RegExp(searchString, 'g'), replaceString);
    fs.writeFileSync(filePath, replacedContent, 'utf8');
};

const copyAndReplace = (sourceDir: string, destinationDir: string, searchString: string, replaceString: string) => {
    if (!fs.existsSync(destinationDir)) {
        fs.mkdirSync(destinationDir);
    } else {
        console.error(`Failed to create component. Directory ${destinationDir} already exists.`);
        process.exitCode = 1;
        return;
    }
    const files = fs.readdirSync(sourceDir);
    files.forEach((file) => {
        const sourceFilePath = path.join(sourceDir, file);
        const destFilePath = path.join(destinationDir, file.replace(new RegExp(searchString), replaceString));
        fs.copyFileSync(sourceFilePath, destFilePath);
        replaceInFile(destFilePath, searchString, replaceString);
    });
};

const generateIndex = () => {
    const componentsDir = path.join(import.meta.dirname, '../src/components');

    const exports = fs
        .readdirSync(componentsDir)
        .filter((component) => component !== 'index.ts' && component !== '__template__' && component !== 'utils.ts')
        .sort()
        .map((name) => {
            return `export * from './${name}';`;
        })
        .join('\n');

    fs.writeFileSync(path.join(componentsDir, 'index.ts'), exports + '\n');
};

const main = () => {
    const args = process.argv.slice(2);
    if (args.length < 1) {
        console.log('Usage: tsx createComponent.ts <componentName>');
        process.exitCode = 1;
        return;
    }
    const componentName = args[0];
    const sourceDir = path.join(import.meta.dirname, '../src/components/__template__');
    const destinationDir = path.join(import.meta.dirname, '../src/components', componentName);
    const searchString = '\\$0';
    const replaceString = componentName;

    copyAndReplace(sourceDir, destinationDir, searchString, replaceString);

    generateIndex();
};

main();
