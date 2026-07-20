# flash-cli

Flash-cli is a tool to create a flashcard deck and view it from your command line. Inspired by the simplicity and power of taskwarrior with a plugin architecture for users to extend it how they wish.

Flashcards are stored in an SQLite database

## Installation
TODO

## Usage
TODO
### Default command
Note on using default command, all arguments passed will be read as filters. ie. a command like `flash-cli group:foo` will review all cards from the `foo` group. But trying to pass the review mode like `flash-cli mode:shuffle` will consider it as a custom filter.
If you want to pass mods to the review command either set defaults or explicitly call `flash-cli review mode:shuffle`

## Plugin creation
Plugin data is stored in a TEXT column as json, so they need to create their own indexes like below

-- run something like this on startup to index its own nested key
CREATE INDEX IF NOT EXISTS idx_plugin1_difficulty 
ON %s (json_extract(ext_data, '$.plugin1.difficulty'));
