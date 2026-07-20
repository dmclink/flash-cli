package ext

import (
	"encoding/json"
	"fmt"
	"time"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/parser"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// toProtoCards converts slice of SQLite database flashcard models into Protobuf wire models
func toProtoCards(dbCards []database.Flashcard) ([]*common.Flashcard, error) {
	protoCards := make([]*common.Flashcard, 0, len(dbCards))
	for _, c := range dbCards {
		pc, err := toProtoCard(c)
		if err != nil {
			return nil, err
		}
		protoCards = append(protoCards, pc)
	}
	return protoCards, nil
}

// toProtoCards converts SQLite database flashcard models into Protobuf wire model
func toProtoCard(c database.Flashcard) (*common.Flashcard, error) {
	var extMap map[string]any
	if len(c.ExtData) > 0 {
		if err := json.Unmarshal(c.ExtData, &extMap); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from database for card ID %d: %w", c.ID, err)
		}
	} else {
		extMap = make(map[string]any)
	}

	extStruct, err := structpb.NewStruct(extMap)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ExtData for card ID %d: %w", c.ID, err)
	}

	return &common.Flashcard{
		Id:         int64(c.ID),
		Uuid:       c.UUID,
		Front:      c.Front,
		Back:       c.Back,
		LastReview: timestamppb.New(time.Unix(int64(c.LastReview), 0)),
		CreatedAt:  timestamppb.New(time.Unix(int64(c.CreatedAt), 0)),
		ExtData:    extStruct,
	}, nil
}

// toModelCards converts slice of incoming cards from a plugin into SQLite flashcard models
func toModelCards(protoCards []*common.Flashcard) ([]database.Flashcard, error) {
	dbCards := make([]database.Flashcard, 0, len(protoCards))
	for _, c := range protoCards {
		dbc, err := toModelCard(c)
		if err != nil {
			return nil, err
		}
		dbCards = append(dbCards, dbc)
	}
	return dbCards, nil
}

// toModelCards converts incoming card from a plugin into SQLite flashcard model
func toModelCard(c *common.Flashcard) (database.Flashcard, error) {
	var extMap map[string]any
	if c.ExtData != nil {
		extMap = c.ExtData.AsMap()
	} else {
		extMap = make(map[string]any)
	}

	extBytes, err := json.Marshal(extMap)
	if err != nil {
		return database.Flashcard{}, fmt.Errorf("marshalling ExtData | %w", err)
	}

	return database.Flashcard{
		ID:         int(c.Id),
		UUID:       c.Uuid,
		Front:      c.Front,
		Back:       c.Back,
		LastReview: int(c.LastReview.AsTime().Unix()),
		CreatedAt:  int(c.CreatedAt.AsTime().Unix()),
		ExtData:    extBytes,
	}, nil
}

func toFilterSet(sf []parser.SearchFilters) common.FilterSet {
	// TODO: SearchFilter has []Filter type
	// need to extract the string value from each
	// need to turn []Filter for Custom into a map[string]string of key value pairs since custom are <key>:<value>
	return common.FilterSet{}
}
