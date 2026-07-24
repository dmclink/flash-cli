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
flash-cli           version             Prints the version
flash-cli           help                Prints help
flash-cli           show                Shows result of all merged configs (including in mem defaults and local overrides)
                                        Accepts a single mod for a search string 
                                        Colorizes overridden keys
flash-cli           summary             Prints summary of cards, counts in groups
flash-cli <filter>  search      <mods>  Search for existing cards
                                        Two modes   1) no mods: use a fzf to look through fronts and backs
                                                    2) with mods produce all cards that match the text input

# Roadmap
- [ ] reviewer interface for native and plugins
    - build out a registry of native reviewer plugins
- [ ] hashicorp/go-plugin for reviewer
    - plugin discovery function
- [ ] regular expression filters
- [ ] config command to set default groups and tags for every add and review
    - using any group or tag in subsequent commands after setting a default overrides the default
    - group, and/or tag, respectively depending which is included
    - ie. `flash-cli config default-group=foo,bar` (command syntax pending)
    - ie. `flash-cli config default-tag=todo` (command syntax pending)
    - `flash-cli add some new::card` will create the new card with groups "foo" and "bar" and the tag "todo"
    - `flash-cli group:baz add another::card` will create a second card only in group "baz", but still include tag "todo"
    - `flash-cli review` will then pull all cards in groups "foo" or "bar" that must have the tag "todo"
    - `flash-cli reset-config default-group default-tag` will be equivalent to `flash-cli config default-group=""` and `flash-cli config default-taag=""`
    - if users want to do a one off review or add without tags this might be bad ux since i have no way to have a single + tag
    - they would need to set default back to ""
    - either reconsider setting default-tags or include a flag like --notag or '+' to mean notag or both
    - currently the parser validation blocks empty tags like '-' and '+' so this would require a change
- [ ] let filters alternatively be split by '=' instead of ':'
- [ ] add `limit:` to filters
    - limit needs to be applied after getting all cards and rearranging for review mode
    - do not simply apply to the SQL get cards query
    - ie. limit 10 on SQL always chooses the top 10 by id
    - review mode is set to shuffle for random draws
    - output would be a random ordering of the top 10 vs expected behavior random pick from all cards 
- [ ] create a map of reserved filters like group: project: limit: etc.
- [ ] for the `add` command have a way to open an editor for easier multi line cards
    - probably just `flash-cli add` without any mods (filters okay) should open it
    - currently using the command like this throws an error
    - might be able to lazily do this by calling add an empty card then calling edit
        - this doesnt leave a way to cancel the add unless i call `delete` if user exits editor without saving
- [ ] change default plugin directory
- [ ] completions for groups, tags, commands, attributes
- [ ] hidden command generate-docs which populates man pages from root command help string
- [ ] go releaser for builds
- [ ] github actions/workflow macOS runners to test if download install works
