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

// UserStore is a persistence structure to get and store users.
type UserStore struct {
	DB *sql.DB
}

// AuthConfig is configuration used for authorization operations
type AuthConfig struct {
	LegacySalt string
	JWTSecret  string
}

func hashPassword(config AuthConfig, password string) string {
	data := []byte(fmt.Sprintf("%v%v", config.LegacySalt, password))
	return fmt.Sprintf("%x", sha1.Sum(data))
}

// Login authenticates user credentials and produces a signed JWT
func (s *UserStore) Login(config AuthConfig, email string, password string) (string, error) {
	hashedPassword := hashPassword(config, password)
	var ID int
	var UUID string
	q := `
		select id, uuid from users where email = ? and password = ?
	`
	err := s.DB.QueryRow(q, email, hashedPassword).Scan(&ID, &UUID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   ID,
		"uuid": UUID,
		"nbf":  time.Now().Unix(),
		"exp":  time.Now().AddDate(0, 0, 5).Unix(),
	})
	return token.SignedString([]byte(config.JWTSecret))
}

// ValidateAndDecodeAuthClaim will ensure the token provided was signed by us and decode its contents
func ValidateAndDecodeAuthClaim(token string, config AuthConfig) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify the algorhythm matches what we original signed
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})
	return t.Claims.(jwt.MapClaims), err
}

// ByID resolves with a user on the channel.
func (s *UserStore) ByID(ID int64) (User, error) {
	user := User{}
	q := `
		select id, email, password, uuid, is_active, reset_password, created, modified
		from users where id = ?
	`
	err := s.DB.QueryRow(q, ID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.UUID,
		&user.IsActive,
		&user.ResetPassword,
		&user.Created,
		&user.Modified)
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
