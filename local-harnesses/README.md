# Useful dummy files and folders for targeting with app configurations

## Usage of config

The config file `build.config.json.template` is designed to have sane defaults for most BloodHound configuration. Copy
it to a new file `build.config.json` and target it for building the app locally.

The config file `integration.config.json.template` is designed to have sane defaults for most BloodHound integration testing configuration.
Copy it to a new file `integration.config.json` and target it for running tests.

This copy is ignored by git, so it will be safe to modify as needed for your local environment needs.

## VS Code

The repository VS Code settings/launch configurations target the above local copies of the files. Once copied, VS Code
will be able to run build and test configurations without any additional setup.
