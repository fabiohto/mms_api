package testutil

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// ExecuteMigrations executa os scripts de migração no banco de dados de teste
func ExecuteMigrations(db *sql.DB, migrationsPath string) error {
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("erro ao ler diretório de migrações: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(migrationsPath, file.Name()))
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo de migração %s: %v", file.Name(), err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("erro ao executar migração %s: %v", file.Name(), err)
		}
	}

	return nil
}

// CleanupDatabase limpa todos os dados das tabelas de teste
func CleanupDatabase(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE mms")
	return err
}

// CreateTestData insere dados de teste no banco de dados
func CreateTestData(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO mms (pair, timestamp, mms20, mms50, mms200) VALUES
		('BRLBTC', NOW(), 150000.0, 148000.0, 145000.0),
		('BRLBTC', NOW() - INTERVAL '1 day', 149000.0, 147000.0, 144000.0),
		('BRLETH', NOW(), 2500.0, 2400.0, 2300.0),
		('BRLETH', NOW() - INTERVAL '1 day', 2400.0, 2300.0, 2200.0)
	`)
	return err
}
