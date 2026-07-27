package cmd

import (
	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/cobra"
)

func NewRootCmd(a *app.App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                constant.APP_NAME,
		Short:              "Flashcard review and management program",
		Long:               "A CLI program to review and manage flashcards. Backed by an SQLite database. Strives for simplicity and ease of use to add and review. Extensible via plugins.",
		DisableFlagParsing: true,
		Version:            constant.VERSION,
	}

	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	// TODO: consider maintaining a global (or context) commands set that gets built here
	// to use for FindCommand in the parser instead of current naive implementation to
	// stop at the first word that doesn't match a filter

	rootCmd.AddCommand(NewVersionCmd(a))
	rootCmd.AddCommand(NewAddCmd(a))
	rootCmd.AddCommand(NewReviewCmd(a))
	rootCmd.AddCommand(NewConfigCmd(a))
	rootCmd.AddCommand(NewEditCmd(a))

	cobra.AddTemplateFunc("maxSubPadding", maxSubPadding)
	cobra.AddTemplateFunc("buildSubSyntax", func(sub *cobra.Command) string {
		if mod, exists := sub.Annotations["modsyntax"]; exists && mod != "" {
			return sub.Name() + " " + mod
		}
		return sub.Name()
	})

	// TODO: delete this line after implementing completions
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.SetUsageTemplate(rootUsageTemplate)
	rootCmd.SetHelpTemplate(rootHelpTemplate)

	return rootCmd
}

func maxSubPadding(cmd *cobra.Command) int {
	maxLen := 0
	for _, sub := range cmd.Root().Commands() {
		if sub.IsAvailableCommand() { // Pure check, no string matching needed
			currentLen := len(sub.Name()) + len(sub.Annotations["modsyntax"])
			if sub.Annotations["modsyntax"] != "" {
				currentLen++
			}
			if currentLen > maxLen {
				maxLen = currentLen
			}
		}
	}
	return maxLen
}

var rootUsageTemplate = `USAGE
{{- if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

COMMANDS
  {{- range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}

Use "{{.CommandPath}} help [command]" for more information about a command.
`

var rootHelpTemplate = `{{.Long}}

USAGE
{{- if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

COMMANDS
{{- range .Commands}}{{if .IsAvailableCommand}}
  {{if .Annotations.filter}}flash-cli <filter> {{else}}flash-cli          {{end -}}
  {{rpad (buildSubSyntax .) (maxSubPadding .)}}   {{.Short}}
{{- end}}{{end}}{{end}}

FILTERS
The <filter> consists of zero or more restrictions on which flashcards to select, such as:
  flash-cli                                      <command> <mods>
  flash-cli 28                                   <command> <mods>
  flash-cli +hard                                <command> <mods>
  flash-cli group:programming                    <command> <mods>
  flash-cli ebeeab00-ccf8-464b-8b58-f7f2d606edfb <command> <mods>
  flash-cli 10-15 group:java,python -easy        <command> <mods>

Some commands may not accept any filters, or only accept certain filters. Those filters
that are skipped may be silently ignored. Check the FILTERS section under the command
in question's help text for any nuance.

By default, all filter elements except tags are combined with an implicit 'or' operator.
Tags are combined with an implicit 'and' operator. Those familiar with taskwarrior syntax
will note this is slightly different behavior.

The purpose for this difference is to enable users to select multiple groups to combine
to review at once. For example the following command:
  task group:java,python -easy review
will grab all flashcards belonging to either the 'java' group OR the 'python' group, 
yet strictly exclude all those with an 'easy' tag.

Filters are parsed as belonging to one of several filter 'types'
  IDs (includes ID Ranges)
  Groups
  + Tags
  - Tags
  UUIDS
  Custom

Groups and Custom filter types are key value pairs that are separated by either a ':' or a '='.
Group filter keys must equal 'group' or match one of its aliases such as 'project' 'grp' 'groups'.
All other filters matching the key value syntax will be considered custom filters.
Custom filters are silently ignored by the core program. However, they are extracted and 
passed through to the review plugin to give third parties options for custom sorting or filtering.

Custom filters keys should only begin with alphabetic letters. Digits or symbols will likely
lead to errors or silent unexpected behavior. For example the following command
  flash-cli -baz:qux review
Will be registered as a negative flag named 'baz:qux' rather than a custom group 
with a key -baz and a value qux.

All filters except UUIDs can be joined with commas as long as they belong to the same 'type'.
Valid:
  flash-cli 1,2-5,19             review   =   flash-cli 1 2 3 4 5 19 review
  flash-cli foo:bar,baz          review   =   flash-cli foo:bar foo:baz review
  flash-cli +hard,veryhard -easy review   =   flash-cli +hard +veryhard -easy review
Invalid
  flash-cli 1,group:foo,10       review

DEFAULT FILTERS
Default filters are applied to most commands unless overridden to reduce repetitive typing
and "batch" adds while working on a project. Notable exceptions such as 'flash-cli edit'
which requires cards explicitly to be targetted don't apply default filters regardless
the config setting. The final filter selection will be deteremined by the manually applied 
filters and the default configs which may be dropped due to one of the manually applied 
filter's precedence. Precedence is determined by the filter type's specificity seen below.

Specifity: [IDs, Ranges, UUIDs] > Groups > Tags.

The reason for this behavior is for users to be able to apply a default group and/or tags 
while adding cards, while occasionally an added card might need a different tag.
A user manually entering different group however is assumed to want no default tags.
A user explicitly targetting an id, uuid, or a range of ids, is assumed to want only those
specific cards that are being targetted.

  CONFIGURATION
    default.filter.groups
	  Comma separated list of default groups.
	  Example: 'default.filter.groups foo,bar'
	default.filter.tags
	  Comma separated list of default tags without prefix. 
	  Only positive '+' tags can be applied.
	  Example: 'default.filter.tags hard,medium'

  EXAMPLES
	The following example commands are not independent, where each preceding commands,
	especially the settings, will affect all following commands.

	ALTERING SETTINGS
	flash-cli config default.filter.groups programming
	  sets the default group to be programming
	flash-cli config default.filter.tags hard
	  sets the default tags to be hard

	ADDING CARDS TO DEFAULT GROUP
	flash-cli add some new::card
	  adds a new card belonging to 'programming' group with a 'hard' tag
	flash-cli add another new::card
	  adds a card to 'programming' group with a 'hard' tag
	flash-cli +easy add some easy::card
	  adds a card to the 'programming' group with an 'easy' tag

	ADDING CARDS TO EXPLICIT GROUP
	flash-cli group:physics add some nerdy::card
	  adds a card to the 'physics' group. Note: default tag not applied because group precedence
	flash-cli group:physics +medium add a tricky nerdy::card
	  adds a card to the 'physics' group. Note: explicit 'medium' tag applied
	flash-cli group:physics +hard add some super nerdy::card
	  adds a card to the 'physics' group. Note: need to explicitly apply 'hard' tag

	REVIEWING CARDS
	flash-cli review
	  starts a review session of all 'programming' cards with 'hard' tag
	  Note: Two cards (the ones with hard tag) will be pulled for review here.
	flash-cli +hard review
	  starts a review session of all 'programming' cards without 'hard' tag
	  Note: Same as above. the 'hard' tag physics card is not pulled for review
	        due to default 'programming' group being applied
	flash-cli -hard review
	  starts a review session of all 'programming' cards without 'hard' tag
	  Note: one card (with the easy tag) will be pulled for review here.
	flash-cli group:physics review
	  starts a review session of all 'physics' cards. Default tag not applied
	  Note: Three cards (all physics cards with any tags) will be pulled for review.
	        The default 'hard' tag is dropped due to explicit group specificity
	flash-cli 1 review
	  only reviews card with id=1 default groups and tags are dropped due to id specificty

	EDITING CARDS
	flash-cli edit
	  throws an error, defaults are ignored and this command needs any explicit tag applied

MODS
The <mods> are any tokens following a valid command. Mods are parsed differently depending
on the individual command that was called. Check under the command in question's help text 
under the MODS section for specific parsing and behavior. Generally, mods are parsed as tokens.
Excess whitespace is collapsed, and certain symbols may cause bash errors. To avoid this, 
mods should be wrapped in single quotes when more than simple single whitespace or symbols
are required.

Since each command is different, 
an exhaustive list won't be provided here aside from three useful examples.

The 'add' command requires 1 or more mod with at least one containing the card separator '::':
	the mods are split at the first instance of card separator '::'
  flash-cli add      this is a new flashcard front::and this is its back
	<mods>        =  'this is a new flashcard front::and this is its back'
	A new card is added where Front='this is a new flashcard' Back='and this is its back'

The 'config' command requires one to two mods:
  The first mod is the config name and the second mod is the new value
  flash-cli config   default.review.limit       10
	<mods>        =  'default.review.limit' '10'
	This will set the default.review.limit config value to 10 in the config file.
  flash-cli config   default.review.limit
	<mods>        =  'default.review.limit'
	Only one mod marks the config key for deletion from the config file.

The 'review' command requires zero or more mods:
  flash-cli review   
	<mods>        =  
	review is ran with default modes and renderers or whichever is set in config
  flash-cli review   mode=shuffle
	<mods>        =  'mode=shuffle'
	review is ran in shuffle mode
  flash-cli review   mode=shuffle renderer=pyrender
	<mods>        =  'mode=shuffle' 'renderer=pyrender'
	review is ran in shuffle mode with pyrender plugin's ui

DEFAULT COMMAND
By default, running this program without any command will run the review command
  flash-cli       =  flash-cli review

Any tokens found after flash-cli will be treated as filters. This is to allow users
to quickly review select groups.
  flash-cli group:programming     will run review on all 'programming' cards

Filter parsing behavior makes it impossible to set an alternate mode on the default command.
  flash-cli mode=shuffle          reviews all cards, with a custom 'mode' filter ignored

Those attributes can be set with 'config' so the default command runs the desired mode.
`

var universalUsageTemplate = `flash-cli {{.Name}} - {{.Short}}

SYNTAX ERROR
  Invalid configuration or syntax provided

USAGE
  flash-cli {{if .Annotations.filter}}<filter> {{else}}         {{end -}}{{.Name}}{{if .Annotations.modsyntax}} <mods>{{end}}

Run 'flash-cli help {{.Name}}' for a detailed list of available options'`
