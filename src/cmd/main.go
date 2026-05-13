package main

import (
	"TEC/src/internal/config"
	"TEC/src/internal/db"
	"TEC/src/internal/session"
	"TEC/src/internal/user"
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	slog.Info("starting TEC server", "port", cfg.Port)

	// 1. БД
	dbConn, err := db.New(cfg)
	if err != nil {
		slog.Error("db init failed", "err", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	// 2. Миграции
	if err := db.AutoMigrate(dbConn); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}

	// 3. Репозиторий пользователей
	userRepo := user.NewRepository(dbConn)

	// 4. Сессии
	sess := session.New()

	// 5. Шаблоны
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		slog.Error("template parse failed", "err", err)
		os.Exit(1)
	}

	// 6. Роутер
	mux := http.NewServeMux()

	// --- ГЛАВНАЯ (перенаправляет на login или dashboard) ---
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sess.Get(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// --- ЛОГИН (GET) ---
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sess.Get(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		tmpl.ExecuteTemplate(w, "login", nil)
	})

	// --- ЛОГИН (POST) ---
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		pass := r.FormValue("password")

		// Ищем пользователя в БД
		u, err := userRepo.GetByEmail(email)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			tmpl.ExecuteTemplate(w, "login", map[string]string{"error": "Invalid credentials"})
			return
		}

		// Проверяем пароль
		if err := u.CheckPassword(pass); err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			tmpl.ExecuteTemplate(w, "login", map[string]string{"error": "Invalid credentials"})
			return
		}

		// Создаем сессию
		sess.Create(w, u.ID)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	// --- РЕГИСТРАЦИЯ (GET) ---
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sess.Get(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		tmpl.ExecuteTemplate(w, "register", nil)
	})

	// --- РЕГИСТРАЦИЯ (POST) ---
	mux.HandleFunc("POST /register", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		pass := r.FormValue("password")
		username := r.FormValue("username")

		newUser := &user.User{
			Email:    email,
			Password: pass,
			Username: username,
		}

		if err := userRepo.Create(newUser); err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register", map[string]string{"error": err.Error()})
			return
		}

		// Автоматический вход после регистрации
		sess.Create(w, newUser.ID)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	// --- ЛОГАУТ ---
	mux.HandleFunc("POST /logout", func(w http.ResponseWriter, r *http.Request) {
		sess.Destroy(w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// --- ДАШБОРД ---
	mux.HandleFunc("GET /dashboard", requireAuth(sess, userRepo, func(w http.ResponseWriter, r *http.Request, u *user.User) {
		data := map[string]any{
			"IsLogin":     false,
			"UserID":      u.ID,
			"Username":    u.Username,
			"Level":       u.Level,
			"XP":          u.XP,
			"NextLevelXP": u.Level * 100, // Простая формула
			"Streak":      u.Streak,
		}

		tmpl.ExecuteTemplate(w, "base", data)
	}))

	// --- MOCK API ---
	mux.HandleFunc("POST /api/generate-roadmap", requireAuth(sess, userRepo, func(w http.ResponseWriter, r *http.Request, u *user.User) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="card"><div class="card-header">GENERATED</div><p style="color:var(--neon-green)">AI Roadmap generated for ` + u.Email + `</p></div>`))
	}))

	// 7. Middleware
	handler := loggingMiddleware(recoveryMiddleware(mux))

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "addr", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen failed", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")
	srv.Shutdown(context.Background())
}

// Middleware: Требует авторизации + загружает пользователя из БД
func requireAuth(sm *session.Manager, repo *user.Repository, next func(http.ResponseWriter, *http.Request, *user.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := sm.Get(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Загружаем свежие данные из БД
		u, err := repo.GetByID(s.UserID)
		if err != nil {
			sm.Destroy(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r, u)
	}
}

// Middleware: Логирование
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}

// Middleware: Восстановление после паники
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
