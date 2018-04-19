package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/oak/model"
	"github.com/urfave/cli"
)

// SQLModel provides a subcommands to work generate structs from existing schema
type SQLModel struct {
	db       *sqlx.DB
	executor *model.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLModel) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "model",
		Usage:        "A group of commands for generating object model from database schema",
		Description:  "A group of commands for generating object model from database schema",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.before,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "schema-name, s",
				Usage: "name of the database schema",
				Value: "",
			},
			cli.StringSliceFlag{
				Name:  "table-name, t",
				Usage: "name of the table in the database",
			},
			cli.StringSliceFlag{
				Name:  "ignore-table-name, i",
				Usage: "name of the table in the database that should be skipped",
				Value: &cli.StringSlice{"migrations"},
			},
			cli.StringFlag{
				Name:  "package-dir, p",
				Usage: "path to the package, where the source code will be generated",
				Value: "./database/model",
			},
			cli.StringFlag{
				Name:  "orm-tag-format, m",
				Usage: "tag format that is wellknow for some ORM packages",
				Value: "sqlx",
			},
			cli.BoolFlag{
				Name:  "keep-schema, k",
				Usage: "keep the schema as package (except default schema)",
			},
			cli.BoolTFlag{
				Name:  "include-docs, d",
				Usage: "include API documentation in generated source code",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:        "print",
				Usage:       "Print the object model for given database schema or tables",
				Description: "Print the object model for given database schema or tables",
				Action:      m.print,
			},
			{
				Name:        "sync",
				Usage:       "Generate a package of models for given database schema",
				Description: "Generate a package of models for given database schema",
				Action:      m.sync,
			},
		},
	}
}

func (m *SQLModel) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return err
	}

	provider, err := m.provider(db)
	if err != nil {
		return err
	}

	builder, err := m.builder(ctx)
	if err != nil {
		return err
	}

	m.db = db
	m.executor = &model.Executor{
		Config: &model.ExecutorConfig{
			KeepSchema: ctx.Bool("keep-schema"),
		},
		Provider: provider,
		Composer: &model.Generator{
			TagBuilder: builder,
			Config: &model.GeneratorConfig{
				KeepSchema:   ctx.Bool("keep-schema"),
				InlcudeDoc:   ctx.BoolT("include-docs"),
				IgnoreTables: ctx.StringSlice("ignore-table-name"),
			},
		},
	}

	return nil
}

func (m *SQLModel) provider(db *sqlx.DB) (model.Provider, error) {
	switch db.DriverName() {
	case "sqlite3":
		return &model.SQLiteProvider{DB: db}, nil
	case "postgres":
		return &model.PostgreSQLProvider{DB: db}, nil
	case "mysql":
		return &model.MySQLProvider{DB: db}, nil
	default:
		err := fmt.Errorf("Cannot find provider for database driver '%s'", db.DriverName())
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}
}

func (m *SQLModel) builder(ctx *cli.Context) (model.TagBuilder, error) {
	builder := model.CompositeTagBuilder{}
	format := ctx.String("orm-tag-format")

	switch strings.ToLower(format) {
	case "sqlx":
		builder = append(builder, model.SQLXTagBuilder{})
	case "gorm":
		builder = append(builder, model.GORMTagBuilder{})
	default:
		err := fmt.Errorf("Cannot find tag builder for '%s'", format)
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}

	builder = append(builder, model.JSONTagBuilder{})
	builder = append(builder, model.XMLTagBuilder{})

	return builder, nil
}

func (m *SQLModel) after(ctx *cli.Context) error {
	if err := m.db.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}
	return nil
}

func (m *SQLModel) print(ctx *cli.Context) error {
	spec := &model.Spec{
		Dir:    ctx.GlobalString("package-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	if err := m.executor.Write(os.Stdout, spec); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLModel) sync(ctx *cli.Context) error {
	spec := &model.Spec{
		Dir:    ctx.GlobalString("package-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	path, err := m.executor.Create(spec)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	if path != "" {
		log.Infof("Generated a database model at: '%s'", path)
	}

	return nil
}