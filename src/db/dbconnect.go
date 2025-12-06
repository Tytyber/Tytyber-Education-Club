package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func InitDB() *sql.DB {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Printf("Предупреждение: .env файл не найден: %v", err)
	}

	// Отладочная информация
	log.Println("=== ПРОВЕРКА ПЕРЕМЕННЫХ ОКРУЖЕНИЯ ===")
	log.Printf("DB_USERNAME: %s", os.Getenv("DB_USERNAME"))
	log.Printf("DB_PASSWORD: %s", os.Getenv("DB_PASSWORD"))
	log.Printf("DB_NAME: %s", os.Getenv("DB_NAME"))
	log.Println("=====================================")

	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbIP := os.Getenv("DB_IP")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
		dbUsername, dbPassword, dbIP, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Ошибка при открытии соединения с базой данных:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("База данных не отвечает:", err)
	}

	log.Println("✅ Успешное подключение к базе данных")
	return db
}
