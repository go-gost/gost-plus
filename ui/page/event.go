package page

type EventID string

const (
	EventThemeChanged EventID = "event.theme.changed"
)

type Event struct {
	ID EventID
}
