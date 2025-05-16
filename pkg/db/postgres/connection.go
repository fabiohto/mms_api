package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// NewConnectionWithTimeout estabelece uma nova conexão com o banco de dados PostgreSQL com timeout
func NewConnectionWithTimeout(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	// Tenta conectar com retry por 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var db *sql.DB
	var err error

	// Tenta conectar a cada 1 segundo
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout ao tentar conectar ao banco de dados: %v", err)
		case <-ticker.C:
			db, err = sql.Open("postgres", connStr)
			if err != nil {
				continue
			}

			// Testa a conexão
			err = db.PingContext(ctx)
			if err == nil {
				// Configura os parâmetros da conexão
				db.SetMaxOpenConns(25)
				db.SetMaxIdleConns(25)
				db.SetConnMaxLifetime(5 * time.Minute)
				return db, nil
			}
		}
	}
}
