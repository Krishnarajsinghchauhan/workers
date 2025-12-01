package internal

type Job struct {
	ID      string            `json:"id"`
	Tool    string            `json:"tool"`
	Files   []string          `json:"files"`
	Options map[string]string `json:"options"`
}
