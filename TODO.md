builtin filters for review should be reviewing cards in certain groups. reviewing cards that last review are too far in the past
Plugins need to be able to extend and allow for their own filters
use -- unix convention to tell command parser to stop looking for filters

flash-cli           groups              Lists all used groups
flash-cli           add         <mods>  Adds a new flashcard, mods is the space separated text for front and back seperated by delimiter
flash-cli           edit                Launches an editor (start with nano or vi) with flashcard info filled in. save buffer pushes updates to db
flash-cli <filter>  modify      <mods>  Overwrites the selected field
flash-cli <filter>  delete              Deletes a flashcard that matches filter
flash-cli <filter>  review      <mods>  Reviews flashcards that match filter, allow mods for reverse review, mixed
                                        Also needs options for review 1 card, all cards in a loop, shuffled, by time last reviewed...
                                        Might need to disallow mods here if we want this to be default which runs without a command
                                        In taskwarrior you can skip the command with `task project:some-project` which will run the default
                                        command `all` and still apply the project filter
flash-cli           config      <mods>  Changes config setting to new value set in mods
                                        Good settings for config are defaults for other commands
                                        ie. review default to mixed, default editor, default groups
flash-cli           version             Prints the version
flash-cli           help                Prints help
flash-cli           summary             Prints summary of cards, counts in groups
flash-cli <filter>  search      <mods>  Search for existing cards
                                        Two modes   1) no mods: use a fzf to look through fronts and backs
                                                    2) with mods produce all cards that match the text input

# Roadmap

- [ ] regular expression filters
