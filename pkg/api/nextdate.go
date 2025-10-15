package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не может быть пустым")
	}

	startTime, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("неверная дата начала: %w", err)
	}

	if repeat == "y" {
		current := startTime
		for {
			current = current.AddDate(1, 0, 0)
			if current.After(now) {
				return current.Format(dateFormat), nil
			}
		}
	} else if after, ok := strings.CutPrefix(repeat, "d "); ok {
		daysStr := after
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return "", fmt.Errorf("недопустимый дневной интервал: %w", err)
		}
		if days <= 0 || days > 400 {
			return "", fmt.Errorf("дневной интервал должен быть от 1 до 400")
		}

		current := startTime
		for {
			current = current.AddDate(0, 0, days)
			if current.After(now) {
				return current.Format(dateFormat), nil
			}

		}
	} else if repeat == "w" || repeat == "m" {
		return "", fmt.Errorf("неподдерживаемый формат")
	} else {
		return "", fmt.Errorf("неизвестный формат правила")
	}
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse("20060202", nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("недействительный now параметр: %v", err), http.StatusBadRequest)
			return
		}
	}

	result, err := NextDate(now, dateStr, repeatStr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, result)
}
