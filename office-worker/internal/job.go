job.gopackage internal

type Job struct {
	ID    string   `json:"id"`
	Tool  string   `json:"tool"`
	Files []string `json:"files"`
}
