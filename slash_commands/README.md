# slash_commands module

## Overview

This module is used to create, update, and delete slash commands from saluki.

[Brief description of slash commands]

Slash commands are stored in the `commands.yml` configuration file.  The YAML is parsed directly into a JSON body, packaged with the saluki Bot authoritization headers, and sent to the Discord API.

## Updating slash commands

### Deployment considerations

## Testing slash commands

`main_test.go` will test attempt to parse `commands.yml` to validate the structure of application commands. 
