package model

// OptionRepository OptionのRepository
type OptionRepository interface {
	InsertOption(lastID int, num int, body string) error
	UpdateOptions(options []string, questionID int) error
	DeleteOptions(questionID int) error
	GetOptions(questionIDs []int) ([]Options, error)
}
