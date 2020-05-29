package main

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/release-bundle-generators/artifactory"
	"github.com/jfrog/release-bundle-generators/docs/common"
	"github.com/jfrog/release-bundle-generators/utils/cliutils"
	"github.com/jfrog/release-bundle-generators/utils/log"
	"github.com/jfrog/jfrog-client-go/utils"
	"os"
)

const commandHelpTemplate string = `{{.HelpName}}{{if .UsageText}}
Arguments:
{{.UsageText}}
{{end}}{{if .VisibleFlags}}
Options:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .ArgsUsage}}
Environment Variables:
{{.ArgsUsage}}{{end}}

`

const appHelpTemplate string = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} [arguments...]{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

const subcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}}[arguments...]

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

func main() {
	log.SetDefaultLogger()
	err := execMain()
	cliutils.ExitOnErr(err)
}

func execMain() error {
	// Set JFrog CLI's user-agent on the jfrog-client-go.
	utils.SetUserAgent(cliutils.GetUserAgent())

	app := cli.NewApp()
	app.Name = "release-bundle-generator"
	app.Usage = "See https://github.com/jfrog/release-bundle-generators for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	app.EnableBashCompletion = true
	app.Commands = artifactory.GetCommands()
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = appHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
	err := app.Run(args)
	return err
}
