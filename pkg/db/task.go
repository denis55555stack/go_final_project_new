package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Функция для добавления задачи
func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить последний идентификатор: %w", err)
	}
	return id, nil
}

// Tasks возвращает список ближайших задач из базы данных.
// limit - максимальное количество возвращаемых записей.
func Tasks(limit int) ([]*Task, error) {

	db := GetDB()
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?` // Запрос с ограничением по количеству

	rows, err := db.Query(query, limit)
	if err != nil {
		log.Printf("ошибка выполнения запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []*Task{}

	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Printf("ошибка сканирования строки: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("ошибка после итерации: %v", err)
		return nil, err
	}

	// возвращаем пустой слайс
	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

var err error

// GetTask возвращает задачу по указанному id.
func GetTask(id int) (*Task, error) {
	db := GetDB()
	err = nil
	if err != nil {
		log.Printf("ошибка подключения к БД: %v", err)
		return nil, err
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	task := &Task{}

	err = db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("задача с id %d не найдена", id)
			return nil, fmt.Errorf("задача с id %d не найдена", id)
		}
		log.Printf("ошибка выполнения запроса: %v", err)
		return nil, err
	}

	return task, nil
}

// UpdateTask обновляет задачу в базе данных.
func UpdateTask(task *Task) error {
	db := GetDB()
	if err != nil {
		log.Printf("ошибка подключения к БД: %v", err)
		return err
	}
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("ошибка выполнения запроса: %v", err)
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Printf("ошибка получения количества измененных строк: %v", err)
		return err
	}

	if count == 0 {
		err := fmt.Errorf("задача с id %s для обновления не найдена", task.ID)
		log.Println(err)
		return err
	}

	return nil
}

// DeleteTask удаляет задачу из базы данных по ID
func DeleteTask(id string) error {
	db := GetDB()
	if err != nil {
		log.Printf("ошибка подключения к БД: %v", err)
		return err
	}

	query := "DELETE FROM scheduler WHERE id = ?"
	_, err = db.Exec(query, id)
	if err != nil {
		log.Printf("не удалось удалить задачу: %v", err)
		return err
	}
	return nil
}

// UpdateDate обновляет только дату у задачи в базе данных.
func UpdateDate(next string, id string) error {
	db := GetDB()
	if err != nil {
		log.Printf("ошибка подключения к БД: %v", err)
		return err
	}

	query := "UPDATE scheduler SET date = ? WHERE id = ?"

	res, err := db.Exec(query, next, id)
	if err != nil {
		log.Printf("ошибка выполнения запроса: %v", err)
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Printf("ошибка получения количества измененных строк: %v", err)
		return err
	}

	if count == 0 {
		err := fmt.Errorf("задача с id %s для обновления не найдена", id)
		log.Println(err)
		return err
	}

	return nil
}
