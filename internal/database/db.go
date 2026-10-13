package database

import (
	"fmt"

	"github.com/darthxd/agendac/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// função para conectar ao banco de dados
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// altera o dialect do banco para o que está definido no .env
	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBDSN)
	case "mysql":
		dialector = mysql.Open(cfg.DBDSN)
	default:
		return nil, fmt.Errorf("DB_DRIVER %q não suportado (use sqlite ou mysql).", cfg.DBDriver)
	}

	logLevel := gormLogger.Warn
	if cfg.Env != "production" {
		logLevel = gormLogger.Info
	}

	// abre a conexão com o banco de dados no dialect escolhido
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:         gormLogger.Default.LogMode(logLevel),
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("abrindo banco (%s): %w", cfg.DBDriver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.DBDriver == "sqlite" {
		sqlDB.SetMaxOpenConns(1)
	}

	// retorna o pointer do gorm.DB
	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
