package ext

import (
	"context"
	"fmt"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	review "github.com/dmclink/flash-cli/gen/go/review/v1"
	"github.com/dmclink/flash-cli/internal/database"
)

type reviewProcessorHostAdapter struct {
	client review.ReviewProcessorServiceClient
}

func (a *reviewProcessorHostAdapter) Process(ctx context.Context, dbCardsIn []database.Flashcard, modifiers []string) ([]database.Flashcard, error) {
	protoCards, err := toProtoCards(dbCardsIn)
	if err != nil {
		return nil, fmt.Errorf("mapping database rows to proto | %w", err)
	}

	protoResp, err := a.client.Process(ctx, &review.ProcessRequest{
		UnparsedModifiers: modifiers,
		Cards:             protoCards,
		// TODO: dont pass empty here, need to parse and convert from parser.SearchFilters, may need to add to signature
		Filters: &common.FilterSet{},
	})
	if err != nil {
		return nil, fmt.Errorf("executing plugin over network | %w", err)
	}

	dbCardsOut, err := toModelCards(protoResp.Cards)
	if err != nil {
		return nil, fmt.Errorf("converting back to database card models | %w", err)
	}

	return dbCardsOut, nil
}

type rendererHostAdapter struct {
	client render.RenderServiceClient
}

func (a *rendererHostAdapter) Init(ctx context.Context) (string, string, string, error) {
	resp, err := a.client.Init(ctx, &render.InitRequest{})
	if err != nil {
		return "", "", "", fmt.Errorf("sending init request | %w", err)
	}

	return resp.StartupBanner, resp.InstructionFront, resp.InstructionBack, nil
}

func (a *rendererHostAdapter) Render(ctx context.Context, cardIn database.Flashcard, cardNum int, cardCount int, modifiers []string) (string, string, string, error) {
	protoCard, err := toProtoCard(cardIn)
	if err != nil {
		return "", "", "", fmt.Errorf("mapping database card to proto | %w", err)
	}

	resp, err := a.client.Process(ctx, &render.ProcessRequest{
		UnparsedModifiers: modifiers,
		// TODO: dont pass empty here, need to parse and convert from parser.SearchFilters, may need to add to signature
		Filters:        &common.FilterSet{},
		Card:           protoCard,
		CurrentCardNum: int32(cardNum),
		TotalCardCount: int32(cardCount),
	})
	if err != nil {
		return "", "", "", fmt.Errorf("executing plugin over network | %w", err)
	}

	return resp.FormattedFront, resp.FormattedBack, resp.Progress, nil
}
