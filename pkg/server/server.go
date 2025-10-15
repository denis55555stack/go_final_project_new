package server

import (
	"fmt"
	"net/http"
)

// Функция для запуска сервера
func Run() error {
	port := 7540
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
