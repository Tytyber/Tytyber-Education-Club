package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
)

func AutoMigrate(db *sql.DB) error {
	slog.Info("starting database migrations...")

	// Проверяем, существует ли таблица миграций
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'schema_migrations'").Scan(&count)
	if err != nil {
		return fmt.Errorf("check migrations table: %w", err)
	}

	// Создаем таблицу версий, если нет
	if count == 0 {
		if _, err := db.Exec("CREATE TABLE schema_migrations (version INT PRIMARY KEY)"); err != nil {
			return fmt.Errorf("create migrations table: %w", err)
		}
	}

	// Читаем файлы из папки migrations
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files) // Сортируем по имени, чтобы порядок был верным

	currentVersion := 0
	db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)

	for _, file := range files {
		// Парсим имя файла: "001_create_users.up.sql" -> версия 1
		var version int
		fmt.Sscanf(filepath.Base(file), "%d_", &version)

		if version <= currentVersion {
			continue // Уже применено
		}

		slog.Info("applying migration", "file", file, "version", version)

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read file %s: %w", file, err)
		}

		// Выполняем SQL внутри транзакции
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert version: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration: %w", err)
		}

		currentVersion = version
	}

	slog.Info("database migrations completed", "current_version", currentVersion)
	return nil
}
