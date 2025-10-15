package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(256) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
); 
CREATE INDEX date_index ON scheduler (date);
`

var db *sql.DB

// Init инициализирует соединение с базой данных и схему.
func Init(dbFile string) error {
	var install bool
	log.Printf("попытка инициализации базы данных: %s", dbFile)

	// Проверяем, существует ли файл базы данных.
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Println("файла базы данных не существует, будет создан.")
		install = true
	} else if err != nil {
		return fmt.Errorf("ошибка проверки файла %v", err)
	}

	var err error
	// Открываем соединение с базой данных.
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия базы данных %v", err)
	}
	log.Println("соединение с базой данных установлено.")

	// Проверяем, работает ли соединение с базой данных.
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к базе данных %v", err)
	}
	log.Println("успешное подключение к базе данных")

	// Устанавливаем схему, если необходимо.
	if install {
		log.Println("создание схемы базы данных...")
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("ошибка создания схемы %v", err)
		}
		log.Println("схема базы данных успешно создана.")
	} else {
		log.Println("схема базы данных уже существует (файл БД найден).")
	}

	return nil
}

// GetDB возвращает соединение с базой данных.
func GetDB() *sql.DB {
	return db
}
