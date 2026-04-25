// Утилита для генерации тестового JWT токена.
// Создаёт пользователя в БД (если не существует) и выдаёт access token.
//
// Запуск:
//   go run ./tools/gentoken
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbURL     = "postgres://aifa:aifa@postgres:5432/aifa?sslmode=disable&search_path=auth,public"
	jwtSecret = "change-me-in-production-min-32-chars-long-secret"
	ttl       = 720 * time.Hour // 30 дней
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping db: %v\n", err)
		os.Exit(1)
	}

	// Найти или создать тестового пользователя
	var userID uuid.UUID
	err = pool.QueryRow(ctx,
		`SELECT id FROM users WHERE auth_provider = 'test' AND provider_id = 'test-user'`,
	).Scan(&userID)

	if err != nil {
		userID = uuid.New()
		_, err = pool.Exec(ctx,
			`INSERT INTO users (id, auth_provider, provider_id, created_at)
			 VALUES ($1, 'test', 'test-user', NOW())
			 ON CONFLICT (auth_provider, provider_id) DO UPDATE SET id = EXCLUDED.id
			 RETURNING id`,
			userID,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert user: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Создан новый тестовый пользователь")
	} else {
		fmt.Println("Найден существующий тестовый пользователь")
	}

	// Генерируем JWT токен — структура должна совпадать с auth-service Claims
	claims := jwt.MapClaims{
		"user_id":    userID.String(),
		"token_type": "access",
		"exp":        time.Now().Add(ttl).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nUser ID:      %s\n", userID)
	fmt.Printf("Access Token: %s\n", signed)
	fmt.Printf("\nДля Swagger — вставь в Authorize:\n  Bearer %s\n", signed)
}
