package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
)

var dateRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func mkdirAllImpl(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func osReadFileImpl(name string) ([]byte, error) { return os.ReadFile(name) }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
