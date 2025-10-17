package main

import (
	"fmt"
	"go1f/pkg/db"
	"log"

	"go1f/pkg/server"

	"github.com/Yandex-Practicum/go_final_project/pkg/api"
)

func main() {

	// Создаем базу данных
	dbFile := "scheduler.db"

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	defer func() {
		if err := db.GetDB().Close(); err != nil {
			log.Printf("Ошибка при закрытии базы данных: %v", err)
		}
	}()

	log.Println("База данных успешно инициализирована.")

	// Иницилизируем Api
	api.Init()

	//Запускаем сервер
	log.Println("Запуск сервера.")
	if err := server.Run(); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v", err)
	}

}
