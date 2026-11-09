package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Регистрация хозяина ----------
func RegisterOwner(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			FullName string `json:"full_name"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

		_, err := db.Exec(`INSERT INTO users (full_name, email, phone, password_hash, role)
                           VALUES ($1, $2, $3, $4, 'owner')`,
			input.FullName, input.Email, input.Phone, string(hashed))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "владелец зарегистрирован успешно"})
	}
}

// ---------- Регистрация няни ----------
func RegisterSitter(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 2️⃣ создаём запись в таблице sitters
		_, err = db.Exec(`
            INSERT INTO sitters (sitter_id, experience_years, certificates, preferences, location, status)
            VALUES ($1, $2, $3, $4, $5, 'pending')
        `, sitterID, input.ExperienceYears, input.Certificates, input.Preferences, input.Location)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "няня зарегистрирована, ожидает подтверждения"})
	}
}

// ---------- Авторизация ----------
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
			return
		}

		var id int
		var role string
		var hash string

		err := db.QueryRow(`SELECT user_id, role, password_hash FROM users WHERE email=$1`, input.Email).
			Scan(&id, &role, &hash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный пароль"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "вход выполнен",
			"role":    role,
		})
	}
}
