package user

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// User - структура пользователя
type User struct {
	ID       int64
	Email    string
	Username string
	Password string // Только при регистрации, не сохраняется в БД
	XP       int
	Level    int
	Streak   int
}

// Repository - работа с БД
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create - регистрация нового пользователя
func (r *Repository) Create(u *User) error {
	// Валидация email
	if !isValidEmail(u.Email) {
		return errors.New("invalid email format")
	}

	// Валидация пароля
	if len(u.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Если username пустой, берем часть email до @
	if u.Username == "" {
		parts := strings.Split(u.Email, "@")
		u.Username = parts[0]
	}

	// Вставка в БД
	query := `
		INSERT INTO users (email, password_hash, username, xp, level, streak)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = r.db.QueryRow(
		query,
		u.Email,
		string(hashedPassword),
		u.Username,
		0, // initial XP
		1, // initial Level
		0, // initial Streak
	).Scan(&u.ID)

	if err != nil {
		// Проверяем на дубликат email
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("email already exists")
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetByEmail - поиск пользователя по email (для логина)
func (r *Repository) GetByEmail(email string) (*User, error) {
	u := &User{}
	var passwordHash string

	query := `
		SELECT id, email, username, password_hash, xp, level, streak
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&passwordHash,
		&u.XP,
		&u.Level,
		&u.Streak,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	u.Password = passwordHash // Сохраняем хэш для проверки
	return u, nil
}

// GetByID - поиск по ID (для сессии)
func (r *Repository) GetByID(id int64) (*User, error) {
	u := &User{}

	query := `
		SELECT id, email, username, xp, level, streak
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.XP,
		&u.Level,
		&u.Streak,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

// CheckPassword - проверка пароля
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Валидация email (простая)
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
