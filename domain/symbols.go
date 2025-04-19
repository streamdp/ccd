package domain

type Symbol struct {
	Symbol  string `db:"symbol"  json:"symbol"`
	Unicode rune   `db:"unicode" json:"unicode"`
}
