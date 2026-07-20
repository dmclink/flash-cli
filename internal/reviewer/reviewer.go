package reviewer

import (
	"time"
)

// TODO: write glue code to turn database.Flashcard into shared.Flashcard
// maybe not here, in reviewer/mapper.go or something
// also need glue code to turn shared.Flashcard back to database.Flashcard, probably

// TODO: probably want to call the plugin map from shared instead of this
// func Processor(mode string) ReviewProcessor {
// 	switch mode {
// 	case SHUFFLE_MODE_KEY:
// 		return reviewer.ShuffleMode{}
// 	case LAST_REVIEW_MODE_KEY:
// 		return reviewer.LastReviewMode{}
// 	case CREATED_AT_MODE_KEY:
// 		return reviewer.CreatedAtMode{}
// 	}
// }

// TODO: move this to a config someday
const REVIEW_TIMEOUT = 5 * time.Second

// each plugin type can have its own file in there like review.go addcard.go

// func Review(cards []database.Flashcard, mode string, unparsedMods []string) error {
// 	// TODO: consider refactoring everything into their plugin calls and moving it all up to ReviewCmd
// 	processor, cleanup, err := ext.DispenseReviewProcessor(mode)
// 	if err != nil {
// 		return fmt.Errorf("dispensing review processor | %w", err)
// 	}
// 	defer cleanup()
//
// 	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// 	defer stop()
//
// 	ctx, cancel := context.WithTimeout(signalCtx, REVIEW_TIMEOUT)
// 	defer cancel()
//
// 	processedCards, err := processor.Process(ctx, cards, unparsedMods)
// 	if err != nil {
// 		return fmt.Errorf("review plugin card process | %w", err)
// 	}
//
// 	// render
//
// 	// alternate terminal buffer since we use clears below, might not be supported
// 	// in early win10 and earlier but that's their problem
// 	fmt.Print("\033[?1049h")
// 	os.Stdout.Sync()
// 	defer func() {
// 		fmt.Print("\033[?1049l")
// 		os.Stdout.Sync()
// 	}()
//
// 	err = utils.ClearScreen()
// 	if err != nil {
// 		return err
// 	}
//
// 	for _, card := range processedCards {
// 		fmt.Println("Front:")
// 		fmt.Println(card.Front)
// 		fmt.Println()
// 		fmt.Println("Press enter to flip card...")
//
// 		_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
// 		if err != nil {
// 			return fmt.Errorf("failed capturing keystore for flipping card | %w", err)
// 		}
//
// 		err = utils.ClearScreen()
// 		if err != nil {
// 			return err
// 		}
//
// 		fmt.Println("Front:")
// 		fmt.Println(card.Front)
// 		fmt.Println()
//
// 		fmt.Println("Back:")
// 		fmt.Println(card.Back)
// 		fmt.Println()
// 		fmt.Println("Press enter for next card...")
//
// 		_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
// 		if err != nil {
// 			return fmt.Errorf("failed capturing keystore for flipping card | %w", err)
// 		}
//
// 		err = utils.ClearScreen()
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }
