package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func toPtrBool(b bool) *bool {
	return &b
}

var CLI struct {
	LogLevel     string       `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"info"`
	LogFormat    string       `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	TemplatesDir string       `help:"Directory containing config templates. Overrides embedded templates." default:""`
	CreateEnv    CreateEnvCmd `cmd:"" help:"Create a new S3C workbench environment."`
	Up           UpCmd        `cmd:"" help:"Start an S3C workbench environment."`
	Configure    ConfigureCmd `cmd:"" help:"Generate configuration files from templates."`
	Destroy      DestroyCmd   `cmd:"" help:"Destroy an S3C workbench environment."`
	Down         DownCmd      `cmd:"" help:"Stop an S3C workbench environment."`
	Logs         LogsCmd      `cmd:"" help:"View logs of an S3C workbench environment."`
}

func main() {
	cmd := kong.Parse(&CLI,
		kong.Name("s3c-workbench"),
		kong.Description("Run a light S3C for development and testing."),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	var logLevel zerolog.Level
	switch CLI.LogLevel {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		panic("invalid log level: " + CLI.LogLevel)
	}

	var writer io.Writer
	switch CLI.LogFormat {
	case "json":
		writer = os.Stdout
	case "text":
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	default:
		panic("invalid log format: " + CLI.LogFormat)
	}

	log.Logger = zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()

	// tmpl, err := workbench.ConfigTemplates.ReadFile("config/backbeat/config.json")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to read embedded config template")
	// }

	// fmt.Println(string(tmpl))

	// ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	// <-ctx.Done()

	err := cmd.Run()
	cmd.FatalIfErrorf(err)

	// cfg, err := LoadConfig("")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to load config")
	// }

	// fmt.Printf("%+v\n", cfg)

	log.Info().Msg("Exiting s3c-workbench")
}
