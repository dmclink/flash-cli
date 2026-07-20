package shared

import (
	addcard "github.com/dmclink/flash-cli/gen/go/addcard/v1"
	review "github.com/dmclink/flash-cli/gen/go/review/v1"
)

// type Flashcard struct {
// 	ID         int
// 	UUID       string
// 	Front      string
// 	Back       string
// 	LastReview time.Time
// 	CreatedAt  time.Time
// 	ExtData    map[string]any
// }

// type FilterSet struct {
// 	Tags    []string
// 	Groups  []string
// 	Ranges  []string
// 	IDs     []string
// 	Customs map[string]string
// }

//
// type ReviewProcessRequest struct {
// 	Filters           FilterSet
// 	UnparsedModifiers []string
// 	Cards             []Flashcard
// }
//
// type ReviewProcessResponse struct {
// 	Flashcards []Flashcard
// }
//
// type AddCardProcessRequest struct {
// 	Card Flashcard
// }
//
// type AddCardProcessResponse struct {
// 	Card Flashcard
// }

type ReviewProcessor interface {
	GenericPluginHandler[*review.ProcessRequest, *review.ProcessResponse]
}

type AddCardProcessor interface {
	GenericPluginHandler[*addcard.ProcessRequest, *addcard.ProcessResponse]
}
