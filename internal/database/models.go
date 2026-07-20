package database

type Flashcard struct {
	ID         int    `db:"id"`
	UUID       string `db:"uuid"`
	LastReview int    `db:"last_review"`
	Front      string `db:"front"`
	Back       string `db:"back"`
	CreatedAt  int    `db:"created_at"`
	ExtData    []byte `db:"ext_data"`
}
