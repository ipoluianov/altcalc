package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HistoryLimit is how many calculations the history keeps
const HistoryLimit = 500

// HistoryEntry is a calculation: the expression as it was typed and the
// result as it reads back into an expression
type HistoryEntry struct {
	Expr   string
	Result string
}

// History is what is kept of the work between the starts: the
// calculations, the variables and the memory, the numbers written as they
// read back into an expression
type History struct {
	Entries []HistoryEntry    `json:",omitempty"`
	Vars    map[string]string `json:",omitempty"`
	Ans     string            `json:",omitempty"`
	Memory  string            `json:",omitempty"`
}

func historyPath() string {
	return filepath.Join(ConfigDirectory(), "history.json")
}

// LoadHistory reads the history; an empty one when there is none
func LoadHistory() History {
	var h History
	if bs, err := os.ReadFile(historyPath()); err == nil {
		_ = json.Unmarshal(bs, &h)
	}
	if len(h.Entries) > HistoryLimit {
		h.Entries = h.Entries[len(h.Entries)-HistoryLimit:]
	}
	return h
}

// SaveHistory writes the history, keeping the last HistoryLimit entries
func SaveHistory(h History) error {
	if len(h.Entries) > HistoryLimit {
		h.Entries = h.Entries[len(h.Entries)-HistoryLimit:]
	}
	bs, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(ConfigDirectory(), 0755); err != nil {
		return err
	}
	return writeFileAtomic(historyPath(), bs)
}
