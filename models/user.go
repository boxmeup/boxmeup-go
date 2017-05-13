package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"crypto/sha1"

	"github.com/dgrijalva/jwt-go"
)

// User is a user entity structure
type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	UUID          string    `json:"uuid"`
	IsActive      bool      `json:"is_active"`
	ResetPassword bool      `json:"-"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

type AuthConfig struct {
	LegacySalt string
	JWTSecret  string
}

func hashPassword(config AuthConfig, password string) string {
	data := []byte(fmt.Sprintf("%v%v", config.LegacySalt, password))
	return fmt.Sprintf("%x", sha1.Sum(data))
}

// Login authenticates user credentials and produces a signed JWT
func Login(db *sql.DB, config AuthConfig, email string, password string) (string, error) {
	hashedPassword := hashPassword(config, password)
	var ID int
	var UUID string
	q := `
		select id, uuid from users where email = ? and password = ?
	`
	err := db.QueryRow(q, email, hashedPassword).Scan(&ID, &UUID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   ID,
		"uuid": UUID,
		// 5 day expiration
		"nbf": time.Now().AddDate(0, 0, 5).Unix(),
	})
	return token.SignedString([]byte(config.JWTSecret))
}

// GetUserByID resolves with a user on the channel.
func GetUserByID(db *sql.DB, ID int) (User, error) {
	user := User{}
	q := `
		select id, email, password, uuid, is_active, reset_password, created, modified
		from users where id = ?
	`
	err := db.QueryRow(q, ID).Scan(&user.ID, &user.Email, &user.Password, &user.UUID, &user.IsActive, &user.ResetPassword, &user.Created, &user.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			// user not found
			// @todo consider sending a custom error that the route handler can consume
		} else {
			log.Fatal(err)
		}
	}
	return user, err
}
