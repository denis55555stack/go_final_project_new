package api

import (
	"net/http"
)

const webDir = "web"

func registerStatic() {
	fileHandler := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileHandler)
}

// Init регистрирует API обработчики.
func Init() {
	registerStatic()
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/tasks", GetTasksHandler)
	http.HandleFunc("/api/task/done", DoneTaskHandler)
}
