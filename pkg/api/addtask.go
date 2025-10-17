package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"go1f/pkg/db"
	"log"
	"net/http"
	"strconv"
	"time"
)

// writeJson — универсальная функция для возврата JSON
func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

// checkDate проверяет и корректирует дату задачи
func checkDate(task *db.Task) error {
	now := time.Now()

	// Если дата не указана сегодняшняя
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Проверяем формат даты
	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		log.Println("некорректный формат даты")
		return errors.New("некорректный формат даты")
	}
	log.Println("Формат даты нормальный")
	// Проверяем правило повторения
	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Println("некорректное правило повторения")
			return fmt.Errorf("error in NextDate: %w", err)
		}

		// Если дата задачи в прошлом берём next
		if !t.After(now) {
			task.Date = next
		} else {
			// Для неповторяющихся задач проверяем, что дата не в прошлом
			if t.Before(now.Truncate(24 * time.Hour)) {
				task.Date = now.Format("20060102")
			}
			return nil
		}
	}

	// Если правило не указано и дата в прошлом сегодняшняя
	if !t.After(now) {
		task.Date = now.Format("20060102")
	}
	log.Println("ошибок нет")

	return nil
}

// TaskHandler обрабатывает запросы Post, Put, Get и Delete для /api/task
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodPut:
		UpdateTaskHandler(w, r)
	case http.MethodGet:
		GetTaskHandlerId(w, r)
	case http.MethodDelete:
		DeleteTaskHandler(w, r)
	default:
		http.Error(w, "неверный метод", http.StatusMethodNotAllowed)
	}
}

// Обработчик для добавления задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "неверный метод", http.StatusMethodNotAllowed)
		return
	}

	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ошибка десериализации JSON: %v", err)
		writeJson(w, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	log.Println("разобрал сообщение. Иду дальше", task)

	// Проверка обязательного поля Title
	if task.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: не указан заголовок задачи")
		writeJson(w, map[string]string{"error": "не указан заголовок задачи"})
		return
	}

	// Проверка даты
	err = checkDate(&task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ошибка проверки даты: %v", err)
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Сохраняем задачу в БД
	id, err := db.AddTask(&task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка при добавлении задачи в БД: %v", err)
		writeJson(w, map[string]string{"error": err.Error()})
		return

	}

	// Возвращаем id
	w.WriteHeader(http.StatusCreated)
	writeJson(w, map[string]any{"id": id})
}

// Обработчик для редактирования задачи

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ошибка десериализации JSON: %v", err)
		writeJson(w, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	// Проверка обязательного поля Title
	if task.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: не указан заголовок задачи")
		writeJson(w, map[string]string{"error": "не указан заголовок задачи"})
		return
	}

	// Проверка идентификатора
	if task.ID == "0" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: не указан идентификатор задачи")
		writeJson(w, map[string]string{"error": "не указан идентификатор задачи"})
		return
	}

	// Проверка даты
	err = checkDate(&task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ошибка проверки даты: %v", err)
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Сохраняем изменения в БД
	err = db.UpdateTask(&task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка при обновлении в базе данных: %v", err)
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, map[string]any{}) // Возвращаем пустой JSON
}

// Обработчик для получения списка задач

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(50) // Максимальное количество записей = 50
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка получения задач из БД: %v", err)
		writeJson(w, map[string]string{"error": "ошибка получения задач из БД"})
		return
	}

	// Создаем пустой слайс
	if tasks == nil {
		tasks = []*db.Task{}
	}

	w.WriteHeader(http.StatusOK)
	writeJson(w, TasksResp{
		Tasks: tasks,
	})

}

// Обработчик для получения задачи по id
func GetTaskHandlerId(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: не указан идентификатор")
		writeJson(w, map[string]string{"error": "не указан идентификатор"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ошибка преобразования идентификатора '%s': %v", idStr, err)
		writeJson(w, map[string]string{"error": "некорректный идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("ошибка получения задачи из БД: %v", err)
		writeJson(w, map[string]string{"error": "задача не найдена"})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJson(w, task)
}

// Обработчик для удаления задач
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: отсутствует параметр id")
		writeJson(w, map[string]string{"error": "отсутствует параметр id"})
		return
	}

	err := db.DeleteTask(idStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка при удалении задачи с id %s: %v", idStr, err)
		writeJson(w, map[string]string{"error": "не удалось удалить задачу"})
		return
	}

	writeJson(w, map[string]any{}) // Пустой JSON

}

// Обработчик для выполненной задачи
func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ошибка: отсутствует параметр id")
		writeJson(w, map[string]string{"error": "отсутствует параметр id"})
		return
	}

	// Преобразуем idStr в int
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка преобразования идентификатора '%s': %v", idStr, err)
		writeJson(w, map[string]string{"error": "Недопустимый параметр идентификатора. Значение должно быть целым числом."})
		return

	}

	// Получить задачу из базы данных
	task, err := db.GetTask(idInt)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("ошибка получения задачи с id %d: %v", idInt, err)
		writeJson(w, map[string]string{"error": "не удалось получить задачу"})
		return
	}
	// Удаляем, если отсутствует правило повторения
	if task.Repeat == "" {
		err = db.DeleteTask(idStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("ошибка удаления задачи с id %s: %v", idStr, err)
			writeJson(w, map[string]string{"error": "не удалось удалить задачу"})

		}

		writeJson(w, map[string]any{})

		return

	}

	nextDay := time.Now().AddDate(0, 0, 1)

	nextDate, err := NextDate(nextDay, task.Date, task.Repeat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка расчета следующей даты: %v", err)
		writeJson(w, map[string]string{"error": "не удалось рассчитать следующую дату."})
		return
	}

	// Обновить дату задачи
	err = db.UpdateDate(nextDate, idStr) // Используем UpdateDate для обновления даты
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ошибка обновления даты задачи с id %s: %v", idStr, err)
		writeJson(w, map[string]string{"error": "не удалось обновить дату задачи"})
		return
	}

	writeJson(w, map[string]any{}) // Пустой JSON

}
