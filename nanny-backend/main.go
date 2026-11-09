package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// loadEnv loads key-value pairs from .env if present without requiring external libraries
func loadEnv() {
	candidates := []string{".env", "../.env", "nanny-backend/.env"}
	for _, path := range candidates {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)
				if os.Getenv(k) == "" {
					os.Setenv(k, v)
				}
			}
		}
		log.Println("📁 Загружен файл окружения:", path)
		break
	}
}

// initDB creates required tables if they don't exist yet
func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		full_name VARCHAR(100),
		email VARCHAR(100) UNIQUE,
		phone VARCHAR(20),
		password_hash VARCHAR(255),
		role VARCHAR(10) CHECK (role IN ('owner', 'sitter', 'admin')),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS pets (
		pet_id SERIAL PRIMARY KEY,
		owner_id INT REFERENCES users(user_id),
		name VARCHAR(50),
		type VARCHAR(20) CHECK (type IN ('cat', 'dog', 'rodent')),
		age INT,
		notes TEXT
	);

	CREATE TABLE IF NOT EXISTS sitters (
		sitter_id INT PRIMARY KEY REFERENCES users(user_id),
		experience_years INT,
		certificates TEXT,
		preferences TEXT,
		location VARCHAR(100),
		status VARCHAR(10) CHECK (status IN ('pending', 'approved', 'rejected'))
	);

	CREATE TABLE IF NOT EXISTS services (
		service_id SERIAL PRIMARY KEY,
		sitter_id INT REFERENCES sitters(sitter_id),
		type VARCHAR(20) CHECK (type IN ('walking', 'boarding', 'home-care')),
		price_per_hour DECIMAL(10,2),
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS bookings (
		booking_id SERIAL PRIMARY KEY,
		owner_id INT REFERENCES users(user_id),
		sitter_id INT REFERENCES sitters(sitter_id),
		pet_id INT REFERENCES pets(pet_id),
		service_id INT REFERENCES services(service_id),
		start_time TIMESTAMP,
		end_time TIMESTAMP,
		status VARCHAR(15) CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed'))
	);

	CREATE TABLE IF NOT EXISTS payments (
		payment_id SERIAL PRIMARY KEY,
		booking_id INT REFERENCES bookings(booking_id),
		amount DECIMAL(10,2),
		method VARCHAR(50),
		status VARCHAR(10) CHECK (status IN ('paid', 'refunded', 'failed')),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS reviews (
		review_id SERIAL PRIMARY KEY,
		booking_id INT REFERENCES bookings(booking_id),
		owner_id INT REFERENCES users(user_id),
		sitter_id INT REFERENCES sitters(sitter_id),
		rating INT CHECK (rating BETWEEN 1 AND 5),
		comment TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS chats (
		chat_id SERIAL PRIMARY KEY,
		booking_id INT REFERENCES bookings(booking_id),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS messages (
		message_id SERIAL PRIMARY KEY,
		chat_id INT REFERENCES chats(chat_id),
		sender_id INT REFERENCES users(user_id),
		content TEXT,
		sent_at TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := db.Exec(schema)
	return err
}

func main() {
	loadEnv()

	// 🔹 Строка подключения к БД (берется из переменной DATABASE_URL, если нет — дефолт)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:Ana4aBada$$@localhost:5432/nanny_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Println("⚠️ Ошибка при вызове sql.Open:", err)
	} else if err := db.Ping(); err != nil {
		log.Println("⚠️ Не удалось подключиться к БД при запуске:", err)
		log.Println("ℹ️ Убедитесь, что переменная DATABASE_URL правильно указана в настройках Render Environment Variables!")
	} else {
		log.Println("✅ Подключение к БД успешно установлено")

		// 🔹 Автоматическое создание таблиц, если они отсутствуют
		if err := initDB(db); err != nil {
			log.Println("⚠️ Предупреждение при инициализации структуры таблиц:", err)
		} else {
			log.Println("✅ Структура таблиц БД готова")
		}

		// Проверка таблицы users
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			log.Println("⚠️ Не удалось получить количество пользователей:", err)
		} else {
			fmt.Println("✅ В таблице users:", count, "записей")
		}
	}

	// Определение директории со статическими файлами
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if _, err := os.Stat("./nanny-backend/static"); err == nil {
			staticDir = "./nanny-backend/static"
		}
	}

	r := mux.NewRouter()
	// 🔹 CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// 🔹 Статические файлы (используем FileServer напрямую)
	staticHandler := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))

	// 🔹 Healthcheck для облачных сервисов
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// 🔹 API эндпоинты
	r.HandleFunc("/register/owner", registerOwner).Methods("POST")
	r.HandleFunc("/register/sitter", registerSitter).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")

	// 🔹 HTML страницы
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
	}).Methods("GET")

	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "register.html"))
	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
	}).Methods("GET")

	// Порт (берется из окружения PORT, по умолчанию 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("✅ Server running on http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// ---------- Регистрация владельца ----------
func registerOwner(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "неверные данные"})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	_, err := db.Exec(`INSERT INTO users (full_name, email, phone, password_hash, role)
                       VALUES ($1, $2, $3, $4, 'owner')`,
		input.FullName, input.Email, input.Phone, string(hashed))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "владелец зарегистрирован успешно"})
}

// ---------- Регистрация няни ----------
func registerSitter(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FullName        string `json:"full_name"`
		Email           string `json:"email"`
		Phone           string `json:"phone"`
		Password        string `json:"password"`
		ExperienceYears int    `json:"experience_years"`
		Certificates    string `json:"certificates"`
		Preferences     string `json:"preferences"`
		Location        string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "неверные данные"})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	// 1️⃣ создаём запись в таблице users
	var sitterID int
	err := db.QueryRow(`
        INSERT INTO users (full_name, email, phone, password_hash, role)
        VALUES ($1, $2, $3, $4, 'sitter')
        RETURNING user_id
    `, input.FullName, input.Email, input.Phone, string(hashed)).Scan(&sitterID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 2️⃣ создаём запись в таблице sitters
	_, err = db.Exec(`
        INSERT INTO sitters (sitter_id, experience_years, certificates, preferences, location, status)
        VALUES ($1, $2, $3, $4, $5, 'pending')
    `, sitterID, input.ExperienceYears, input.Certificates, input.Preferences, input.Location)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "няня зарегистрирована, ожидает подтверждения"})
}

// ---------- Авторизация ----------
func login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "неверные данные"})
		return
	}

	var id int
	var role string
	var hash string

	err := db.QueryRow(`SELECT user_id, role, password_hash FROM users WHERE email=$1`, input.Email).
		Scan(&id, &role, &hash)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "пользователь не найден"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный пароль"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "вход выполнен",
		"role":    role,
	})
}
