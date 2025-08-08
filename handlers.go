package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type habit struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

var (
	habitsMap  = map[int]string{}
	habitsList = []habit{}
)

func loadHabits(path string) {
	file := getenv("HABITS_FILE", path)
	b, err := osReadFile(file)
	if err != nil {
		// Defaults with integer codes
		defaults := []habit{{1, "Пить воду"}, {2, "Читать 20 минут"}, {3, "Прогулка 30 минут"}, {4, "Медитация 10 минут"}, {5, "Без кофе сегодня"}}
		habitsMap = map[int]string{}
		habitsList = make([]habit, 0, len(defaults))
		for _, h := range defaults {
			habitsMap[h.Code] = h.Name
			habitsList = append(habitsList, h)
		}
		return
	}
	lines := strings.Split(string(b), "\n")
	habitsMap = map[int]string{}
	habitsList = make([]habit, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue
		}
		codeStr := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		if codeStr == "" || name == "" {
			continue
		}
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			continue
		}
		if _, exists := habitsMap[code]; exists {
			continue
		}
		habitsMap[code] = name
		habitsList = append(habitsList, habit{Code: code, Name: name})
	}
	if len(habitsList) == 0 {
		defaults := []habit{{1, "Привычка"}}
		habitsMap[1] = "Привычка"
		habitsList = defaults
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

func handleHabits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(habitsList)
}

func handleGetMarks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	year := r.URL.Query().Get("year")
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}
	if len(year) != 4 || strings.ContainsAny(year, "^/\\:;,'\" ") {
		http.Error(w, "bad year", http.StatusBadRequest)
		return
	}
	habitStr := r.URL.Query().Get("habit")
	hid, err := strconv.Atoi(habitStr)
	if err != nil {
		http.Error(w, "bad habit", http.StatusBadRequest)
		return
	}
	if _, ok := habitsMap[hid]; !ok {
		http.Error(w, "unknown habit", http.StatusBadRequest)
		return
	}
	// Build epoch range: [year-01-01, nextYear-01-01)
	yi, _ := strconv.Atoi(year)
	nextYear := fmt.Sprintf("%d", yi+1)
	rows, err := db.Query(`
		SELECT strftime('%Y-%m-%d', date, 'unixepoch') AS d
		FROM marks
		WHERE habit = ?
		  AND date >= strftime('%s', ? || '-01-01')
		  AND date <  strftime('%s', ? || '-01-01')
		ORDER BY date`, hid, year, nextYear)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	dates := make([]string, 0)
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dates)
}

type toggleReq struct {
	Date  string `json:"date"`
	Habit int    `json:"habit"`
}

type toggleResp struct {
	Marked bool `json:"marked"`
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// limit request body to small size
	r.Body = http.MaxBytesReader(w, r.Body, 2<<10) // 2KB
	var req toggleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if !dateRe.MatchString(req.Date) {
		http.Error(w, "bad date", 400)
		return
	}
	if _, ok := habitsMap[req.Habit]; !ok {
		http.Error(w, "unknown habit", 400)
		return
	}
	// Toggle: if exists -> delete, else -> insert
	var exists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM marks WHERE habit=? AND date = strftime('%s', ?))`, req.Habit, req.Date).Scan(&exists); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if exists {
		if _, err := db.Exec(`DELETE FROM marks WHERE habit=? AND date = strftime('%s', ?)`, req.Habit, req.Date); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, toggleResp{Marked: false})
		return
	}
	if _, err := db.Exec(`INSERT INTO marks(habit, date) VALUES (?, strftime('%s', ?))`, req.Habit, req.Date); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, toggleResp{Marked: true})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := db.Ping(); err != nil {
		http.Error(w, "db not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleFavicon serves a simple SVG calendar icon with today's date
func handleFavicon(w http.ResponseWriter, r *http.Request) {
	// Generate SVG with current day number
	day := time.Now().Day()
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#ff5f6d"/>
      <stop offset="100%" stop-color="#ffc371"/>
    </linearGradient>
  </defs>
  <!-- calendar body -->
  <rect x="6" y="8" width="52" height="50" rx="8" ry="8" fill="#ffffff" stroke="#e5e5e5"/>
  <!-- header bar -->
  <rect x="6" y="8" width="52" height="14" rx="8" ry="8" fill="url(#g)"/>
  <!-- rings -->
  <circle cx="22" cy="15" r="2.2" fill="#ffffff"/>
  <circle cx="42" cy="15" r="2.2" fill="#ffffff"/>
  <!-- day number -->
  <text x="50%" y="46" text-anchor="middle" font-family="-apple-system,Segoe UI,Roboto,Arial,sans-serif" font-size="28" font-weight="700" fill="#222">%d</text>
</svg>`, day)
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(svg))
}

// a tiny indirection for easier testing
var osReadFile = func(name string) ([]byte, error) { return osReadFileImpl(name) }
