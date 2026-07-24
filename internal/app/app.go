package app

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/config"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/logger"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

type App struct {
	DB     *sql.DB
	Config *config.Config
	Args   parser.ParsedArgs
	Logger hclog.Logger
}

func NewApp(args parser.ParsedArgs) (*App, error) {
	db, err := database.Open()
	if err != nil {
		return nil, fmt.Errorf("finding path and opening database | %v", err)
	}

	err = database.Init(db)
	if err != nil {
		return nil, fmt.Errorf("initializing database tables | %v\n", err)
	}

	cfg, err := config.InitConfig()
	if err != nil {
		return nil, fmt.Errorf("initializing viper config | %w", err)
	}

	logger, err := logger.InitPluginLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing plugin logger | %v\n", err)
	}
	return &App{DB: db, Config: cfg, Args: args, Logger: logger}, nil
}

func (a *App) Close() {
	a.DB.Close()
}

type Command struct {
	*cobra.Command
}
