const core = require('@actions/core');
const github = require('@actions/github');
const fs = require('fs');

try {
  // `version-file` input defined in action metadata file
  const pathToVersionFile = core.getInput('version-file');
  console.log(`file to read ${pathToVersionFile}!`);
  const contents = fs.readFileSync(`${pathToVersionFile}`, 'utf8');
  core.setOutput("version", contents);
  // Get the JSON webhook payload for the event that triggered the workflow
  const payload = JSON.stringify(github.context.payload, undefined, 2)
  console.log(`The event payload: ${payload}`);
} catch (error) {
  core.setFailed(error.message);
}
