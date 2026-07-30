package edit

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/config"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/platform"
	"github.com/dmclink/flash-cli/internal/utils"
)

const INDENT = 21

var template = `# vim: set expandtab tabstop=2 shiftwidth=2 :
# The 'flash-cli <id> edit' command allows you to modify all aspects of a card
# using a text editor. Below is a representation of all the task details.
# Modify what you wish, and when you save and quit your editor,
# The program will read this file, determine what changed, and apply
# those changes. If you exit your editor without saving or making
# modifications, no changes will be made.
#
# Lines that begin with # represent data you cannot change, like ID.
# Edits to these lines will be ignored. The program will attempt to detect
# malformed edits and re-open the file with original content and an error 
# message displayed, no guarantees though. If you get stuck in a loop
# with the same file, just quit without saving.
#
# Do not reorder rows. Entered data must align, only data from 22nd column 
# and after is parsed. Do not edit row Name. Your editor may automatically 
# replace single spaces with tabs. This can result in subtle errors with 
# incorrect row names or truncated data as the data will visually appear 
# to start at the correct column, but actually it is further left. Verify 
# your cursor's column index in the editor is correct.
# 
#
# Name               Editable details
# -----------------  ----------------------------------------------------
# ID:                %s
# UUID:              %s
# Editing Error:     %s
# -----------------  ----------------------------------------------------
# Separate the tags and groups with spaces like this: tag1 tag2
# Names must not start with a number or contain symbols like + ! ,
  Groups:            %s
  Tags:              %s
  Created:           %s
  Last Review:       %s
# For multiline data for Front and Back, new lines must start on the same column
# That is they must be preceded by at least 21 whitespaces
  Front:             %s
  Back:              %s
# Below is the raw JSON data. Edit carefully. Must remain valid JSON.
  Ext Data:          %s
# Flip this to true to break out of the batch editing loop early
  Quit Batch Edit:   false
# END`

type FormattedCard struct {
	ID         string
	UUID       string
	Groups     string
	Tags       string
	CreatedAt  string
	LastReview string
	Front      string
	Back       string
	ExtData    string
}

func (c FormattedCard) toFullFlashcard() (database.FullFlashcard, error) {
	zero := database.FullFlashcard{}
	id, _ := strconv.Atoi(c.ID)

	createdTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(c.CreatedAt), time.Local)
	if err != nil {
		return zero, fmt.Errorf("created field must be in format: YYYY-MM-DD hh-mm-ss")
	}
	created := int(createdTime.Unix())

	lastReviewTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(c.LastReview), time.Local)
	if err != nil {
		return zero, fmt.Errorf("last review field must be in format: YYYY-MM-DD hh-mm-ss")
	}
	lastReview := int(lastReviewTime.Unix())

	extData := []byte(c.ExtData)
	jsonProcessed := map[string]any{}
	err = json.Unmarshal(extData, &jsonProcessed)
	if err != nil {
		return zero, fmt.Errorf("invalid json passed to Ext Data row:\n%s\ncheck structure and quotations", string(extData))
	}

	return database.FullFlashcard{
		Flashcard: database.Flashcard{
			ID:         id,
			UUID:       c.UUID,
			LastReview: lastReview,
			CreatedAt:  created,
			ExtData:    extData,
			Front:      c.Front,
			Back:       c.Back,
		},
		Groups: utils.SplitFieldsAndCommas(c.Groups),
		Tags:   utils.SplitFieldsAndCommas(c.Tags),
	}, nil
}

func FormatCard(db *sql.DB, card database.Flashcard) (FormattedCard, error) {
	zero := FormattedCard{}
	groups, err := database.GetFlashcardGroups(db, card.ID)
	if err != nil {
		fmt.Println(err)
		return zero, err
	}
	tags, err := database.GetFlashcardTags(db, card.ID)
	if err != nil {
		fmt.Println(err)
		return zero, err
	}

	return FormattedCard{
		ID:         strconv.Itoa(card.ID),
		UUID:       card.UUID,
		Groups:     strings.Join(groups, " "),
		Tags:       strings.Join(tags, " "),
		CreatedAt:  time.Unix(int64(card.CreatedAt), 0).Format("2006-01-02 15:04:05"),
		LastReview: time.Unix(int64(card.LastReview), 0).Format("2006-01-02 15:04:05"),
		Front:      indentMultiLine(card.Front, false),
		Back:       indentMultiLine(card.Back, false),
		ExtData:    indentedExtData(string(card.ExtData)),
	}, nil
}

func Edit(ctx context.Context, a *app.App, cards []database.Flashcard) error {
	editorFull := a.Config.Resolve(config.KeyDefaultEditor, "", platform.EditorFallback())
	editorFields := strings.Fields(editorFull)
	editor := editorFields[0]
	editorFlags := editorFields[1:]

	editingError := ""
	cardIdx := 0
batchLoop:
	for cardIdx < len(cards) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		card := cards[cardIdx]
		tmpFile, err := os.CreateTemp("", "flash-cli.*.task")
		if err != nil {
			return err
		}

		cleanup := func(err error) error {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return err
		}

		c, err := FormatCard(a.DB, card)
		if err != nil {
			return cleanup(err)
		}

		text := fmt.Sprintf(template, c.ID, c.UUID, editingError, c.Groups, c.Tags, c.CreatedAt, c.LastReview, c.Front, c.Back, c.ExtData)
		tmpFile.Write([]byte(text))

		bytesBefore, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return cleanup(err)
		}

		err = openEditor(ctx, editor, editorFlags, tmpFile)
		if err != nil {
			return cleanup(err)
		}

		bytesAfter, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return cleanup(err)
		}

		tmpFile.Close()
		os.Remove(tmpFile.Name())
		if bytes.Equal(bytesBefore, bytesAfter) {
			cardIdx++
			continue
		}

		fields := extractFields(bytesAfter)
		err = validateFieldNames(fields)
		if err != nil {
			fmt.Println("Error: Malformed editor input. Restoring data and trying again")
			editingError = indentMultiLine(err.Error(), true)
			continue batchLoop
		}

		if exit := fields["Quit Batch Edit"]; strings.ToLower(exit) == "true" {
			break batchLoop
		}

		newCard, err := toFullFlashcard(card.ID, card.UUID, fields)
		if err != nil {
			fmt.Println("Error: Malformed editor input. Restoring data and trying again")
			editingError = indentMultiLine(err.Error(), true)
			continue batchLoop
		}

		// check for match aginst input even though we checked bytes before
		// some formatting happens in extract fields like whitespace trimming etc.
		// and we dont want to trigger an sql tx for an accidental added whitespace
		fieldsBefore := extractFields(bytesBefore)
		oldCard, _ := toFullFlashcard(card.ID, card.UUID, fieldsBefore)
		if reflect.DeepEqual(newCard, oldCard) {
			cardIdx++
			continue
		}

		err = database.UpdateFlashcard(ctx, a.DB, oldCard, newCard)
		if err != nil {
			fmt.Println("Error: Malformed editor input. Restoring data and trying again")
			editingError = indentMultiLine(err.Error(), true)
			continue batchLoop
			return fmt.Errorf("updating card\nBefore:\n%v\nAfter:\n%v\n\n%w", err)
		}

		cardIdx++
	}

	return nil
}

// Preconditions: all expected fields keys have already been verified exist
func toFullFlashcard(id int, uuid string, fields map[string]string) (database.FullFlashcard, error) {
	zero := database.FullFlashcard{}
	groupsRaw := fields["Groups"]
	groups := utils.SplitFieldsAndCommas(groupsRaw)
	for _, group := range groups {
		if !unicode.IsLetter(rune(group[0])) {
			return zero, fmt.Errorf("group '%s' must start with a letter", group)
		}
	}

	tagsRaw := fields["Tags"]
	tags := utils.SplitFieldsAndCommas(tagsRaw)
	for _, tag := range tags {
		if !unicode.IsLetter(rune(tag[0])) {
			return zero, fmt.Errorf("tag '%s' must start with a letter", tag)
		}
	}

	createdRaw := fields["Created"]
	createdTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(createdRaw), time.Local)
	if err != nil {
		return zero, fmt.Errorf("created field must be in format: YYYY-MM-DD hh-mm-ss")
	}
	created := int(createdTime.Unix())

	lastReviewRaw := fields["Last Review"]
	lastReviewTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(lastReviewRaw), time.Local)
	if err != nil {
		return zero, fmt.Errorf("last review field must be in format: YYYY-MM-DD hh-mm-ss")
	}
	lastReview := int(lastReviewTime.Unix())

	frontRaw := fields["Front"]
	backRaw := fields["Back"]
	extDataRaw := fields["Ext Data"]
	extData := []byte(extDataRaw)

	jsonProcessed := map[string]any{}
	err = json.Unmarshal(extData, &jsonProcessed)
	if err != nil {
		return zero, fmt.Errorf("invalid json passed to Ext Data row:\n%s\ncheck structure and quotations", extDataRaw)
	}

	quitEditRaw := fields["Quit Batch Edit"]
	quitEdit := strings.ToLower(quitEditRaw)
	if quitEdit != "false" && quitEdit != "true" {
		return zero, fmt.Errorf("Quit Batch Edit row has invalid data: '%s'; did you mean to set it to 'true'?", quitEditRaw)
	}

	return database.FullFlashcard{
		Flashcard: database.Flashcard{
			ID:         id,
			UUID:       uuid,
			LastReview: lastReview,
			Front:      frontRaw,
			Back:       backRaw,
			CreatedAt:  created,
			ExtData:    extData,
		},
		Groups: groups,
		Tags:   tags,
	}, nil
}

func extractFields(data []byte) map[string]string {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	lines = stripCommentsFromLines(lines)
	lines = stripBlankBeforeFirstName(lines)
	curr := make([]string, 0, len(lines))
	fields := make(map[string]string, len(lines)*2)

	prevName, firstVal := splitLine(lines[0])
	curr = append(curr, firstVal)
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		name, val := splitLine(line)
		if name != "" {
			fields[prevName] = strings.Join(curr, "\n")
			curr = make([]string, 0, len(lines))
			prevName = name
		}
		curr = append(curr, val)
	}
	fields[prevName] = strings.Join(curr, "\n")

	return fields
}

// validateFieldNames checks fields keys to ensure they contain exactly a predetermined list of names
// returns error if any fields names do not exactly match, or if an extra name exists
func validateFieldNames(fields map[string]string) error {
	wantRowNames := []string{"Groups", "Tags", "Created", "Last Review", "Front", "Back", "Ext Data", "Quit Batch Edit"}

	if len(wantRowNames) > len(fields) {
		return fmt.Errorf("Missing row name.\nCareful deleting rows.")
	}

	if len(wantRowNames) < len(fields) {
		return fmt.Errorf("Extra row name found.\nAdd additional info to Ext Data instead.")
	}

	for _, name := range wantRowNames {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("Cannot find row name '%s'.\nCheck spelling of existing rows and avoid changing their text.", name)
		}
	}

	return nil
}

// getNameFromLine extracts the row name from columns [3,21] (1 indexed) of the row
// If the line length is less than 3 returns an empty string.
// Also strips any semicolons from the end of the vaalue
//
// Preconditions:
// - proper prefix whitespace (two spaces)
// - proper suffix whitespace (total of 21 columns before data starts)
// - no commented lines
func getNameFromLine(line string) string {
	result := strings.TrimSpace(line[min(2, len(line)):min(INDENT, len(line))])
	needsCut := true
	for needsCut {
		result, needsCut = strings.CutSuffix(result, ":")
	}
	return result
}

// getValueFromLine extracts the value of the line (the data from column [22...]).
// If the line length is less than 22 returns an empty string.
// Trims only right whitespace, while maintaining left whitespace for multiline indented text
//
// Preconditions:
// - proper prefix spacing (total of 21 columns before data starts)
// - no commented lines
func getValueFromLine(line string) string {
	if len(line) < INDENT {
		return ""
	}
	return strings.TrimRight(line[INDENT:], " \n\t")
}

// splitLine splits the line into its name (columns [2,21]) and its value (columns [22...])
func splitLine(line string) (string, string) {
	return getNameFromLine(line), getValueFromLine(line)
}

// stripCommentsFromLines remove all lines beginning with '#' strictly at the first column
func stripCommentsFromLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		result = append(result, line)
	}

	return result
}

// stripBlankBeforeFirstName removes all empty lines before the first non-empty Name row
//
// Preconditions: no comment lines (beginning with '#') exist in lines
func stripBlankBeforeFirstName(lines []string) []string {
	result := []string{}
	firstNameFound := false
	for _, line := range lines {
		name := getNameFromLine(line)
		if name != "" {
			firstNameFound = true
		}
		if firstNameFound {
			result = append(result, line)
		}
	}
	return result
}

func openEditor(ctx context.Context, editor string, editorFlags []string, buffer *os.File) error {
	c := exec.CommandContext(ctx, editor, append(editorFlags, buffer.Name())...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func indentedExtData(rawJSON string) string {
	if rawJSON == "" {
		rawJSON = "{}"
	}

	var data interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return "{}"
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "{}"
	}

	return indentMultiLine(string(pretty), false)
}

func indentMultiLine(s string, commented bool) string {
	var indent string
	if commented {
		indent = "#" + strings.Repeat(" ", INDENT-1)
	} else {
		indent = strings.Repeat(" ", INDENT)
	}
	lines := strings.Split(s, "\n")
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
