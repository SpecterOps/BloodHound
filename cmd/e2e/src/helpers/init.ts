import * as fs from 'node:fs';

function createsupportfolders(path: string): void {
    if (!fs.existsSync(path)) {
        fs.mkdirSync(path, { recursive: true });
        console.log(`${path} folder is created`)
    } else {
        console.log(`${path} folder exists`)
    }
}

function createRerunFile(fileName: string): void {
    if (!fs.existsSync(fileName)) {
        fs.writeFileSync(fileName, '');
        console.log(`${fileName} is created`)
    } else {
        console.log(`${fileName} file exists`)
    }
}

const reRunFile = "@rerun.txt";
createRerunFile(reRunFile);

const reportPath = "./test-results/reports"
createsupportfolders(reportPath);

const screenshotsPath = "./test-results/screenshots"
createsupportfolders(screenshotsPath);