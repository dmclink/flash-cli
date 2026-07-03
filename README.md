# flash-cli

Flash-cli is a tool to create a flashcard deck and view it from your command line. Inspired by the simplicity and power of taskwarrior with a plugin architecture for users to extend it how they wish.

Flashcards are stored in an SQLite database

## Installation
TODO

## Usage
TODO

## Plugin creation
Plugin data is stored in a TEXT column as json, so they need to create their own indexes like below

-- run something like this on startup to index its own nested key
CREATE INDEX IF NOT EXISTS idx_plugin1_difficulty 
ON %s (json_extract(ext_data, '$.plugin1.difficulty'));
