package users

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Store is a persistence structure to get and store users.
type Store struct {
	DB *sql.DB
}

// NewStore constructs a storage interface for users.
func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
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
func (s *Store) Login(config AuthConfig, email string, password string) (string, error) {
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

// Register creates a new user in the system.
// @todo Replace shitty password hashing with a more robust mechanism (bcrypt)
func (s *Store) Register(config AuthConfig, email string, password string) (id int64, err error) {
	if s.doesUserExistByEmail(email) {
		return 0, errors.New("user already exists with given email")
	}
	hashedPassword := hashPassword(config, password)
	q := `
		insert into users (email, password, uuid, created, modified)
		values (?, ?, uuid(), now(), now())
	`
	res, err := s.DB.Exec(q, email, hashedPassword)
	id, _ = res.LastInsertId()
	return
}

func (s *Store) doesUserExistByEmail(email string) bool {
	// flesh this out
	q := "select count(*) from users where email = ?"
	var count int
	s.DB.QueryRow(q, email).Scan(&count)
	return count > 0
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
func (s *Store) ByID(ID int64) (User, error) {
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
