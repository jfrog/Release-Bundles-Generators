package artifactory

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/release-bundle-generators/artifactory/commands"
	"github.com/jfrog/release-bundle-generators/artifactory/commands/distribution"
	"github.com/jfrog/release-bundle-generators/docs/artifactory/releasebundletranslatechart"
	"github.com/jfrog/release-bundle-generators/docs/artifactory/use"
	"github.com/jfrog/release-bundle-generators/docs/common"
	"github.com/jfrog/release-bundle-generators/utils/cliutils"
	"github.com/jfrog/release-bundle-generators/utils/config"
	configdocs "github.com/jfrog/release-bundle-generators/docs/artifactory/config"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "config",
			Flags:        getConfigFlags(),
			Aliases:      []string{"c"},
			Usage:        configdocs.Description,
			HelpName:     common.CreateUsage("config", configdocs.Description, configdocs.Usage),
			UsageText:    configdocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: common.CreateBashCompletionFunc("show", "delete", "clear", "import", "export"),
			Action: func(c *cli.Context) error {
				return configCmd(c)
			},
		},
		{
			Name:         "use",
			Usage:        use.Description,
			HelpName:     common.CreateUsage("use", use.Description, use.Usage),
			UsageText:    use.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: common.CreateBashCompletionFunc(commands.GetAllArtifactoryServerIds()...),
			Action: func(c *cli.Context) error {
				return useCmd(c)
			},
		},
		{
			Name:         "from-chart",
			Flags:        getReleaseBundleTranslateChartFlags(),
			Aliases:      []string{"fc"},
			Usage:        releasebundletranslatechart.Description,
			HelpName:     common.CreateUsage("fc", releasebundletranslatechart.Description, releasebundletranslatechart.Usage),
			UsageText:    releasebundletranslatechart.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleTranslateChartCmd(c)
			},
		},
	}
}

func getUrlFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Artifactory URL.` `",
		},
		cli.StringFlag{
			Name:  "dist-url",
			Usage: "[Optional] Distribution URL.` `",
		},
	}
}

func getBaseFlags() []cli.Flag {
	return append(getUrlFlags(),
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Artifactory username.` `",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Artifactory password.` `",
		},
		cli.StringFlag{
			Name:  "apikey",
			Usage: "[Optional] Artifactory API key.` `",
		},
		cli.StringFlag{
			Name:  "access-token",
			Usage: "[Optional] Artifactory access token.` `",
		})
}

func getInsecureTlsFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "insecure-tls",
		Usage: "[Default: false] Set to true to skip TLS certificates verification.` `",
	}
}

func getClientCertsFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "client-cert-path",
			Usage: "[Optional] Client certificate file in PEM format.` `",
		},
		cli.StringFlag{
			Name:  "client-cert-key-path",
			Usage: "[Optional] Private key file for the client certificate in PEM format.` `",
		},
	}
}

func getCommonFlags() []cli.Flag {
	flags := append(getBaseFlags(),
		cli.StringFlag{
			Name:  "ssh-passphrase",
			Usage: "[Optional] SSH key passphrase.` `",
		})
	return append(flags, getSshKeyPathFlag()...)
}

func getServerFlags() []cli.Flag {
	return append(getCommonFlags(), getServerIdFlag())
}

func getServerIdFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "server-id",
		Usage: "[Optional] Artifactory server ID configured using the JFrog CLI config command.` `",
	}
}

func getSshKeyPathFlag() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "ssh-key-path",
			Usage: "[Optional] SSH key file path.` `",
		},
	}
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolTFlag{
			Name:  "interactive",
			Usage: "[Default: true, unless $CI is true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.` `",
		},
		cli.BoolTFlag{
			Name:  "enc-password",
			Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifactory's encryption API.` `",
		},
	}
	flags = append(flags, getBaseFlags()...)
	flags = append(flags, getClientCertsFlags()...)
	return append(flags,
		getSshKeyPathFlag()...)
}

func getDistributionPassphraseFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "passphrase",
		Usage: "[Optional] The passphrase for the signing key. ` `",
	}
}

func getStoringRepositoryFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "repo",
		Usage: "[Optional] A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one.` `",
	}
}

func getReleaseBundleTranslateChartFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name: "chart-path",
			Usage: "Path to a Helm chart in Artifactory, which should be used to generate a release bundle.` `",
		},
		cli.StringFlag{
			Name: "docker-repo",
			Usage: "A Docker repository containing all the Docker images the Helm chart requires.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with JFrog Distribution.` `",
		},
		cli.BoolFlag{
			Name:  "sign",
			Usage: "[Default: false] If set to true, automatically signs the release bundle version.` `",
		},
		cli.StringFlag{
			Name:  "desc",
			Usage: "[Optional] Description of the release bundle.` `",
		},
		cli.StringFlag{
			Name:  "release-notes-path",
			Usage: "[Optional] Path to a file describes the release notes for the release bundle version.` `",
		},
		cli.StringFlag{
			Name:  "release-notes-syntax",
			Usage: "[Default: plain_text] The syntax for the release notes. Can be one of 'markdown', 'asciidoc', or 'plain_text` `",
		},
		cli.StringFlag{
			Name:  "exclusions",
			Usage: "[Optional] Semicolon-separated list of exclusions. Exclusions can include the * and the ? wildcards.` `",
		},
		getDistributionPassphraseFlag(),
		getStoringRepositoryFlag(),
	}...)
}

func createArtifactoryDetailsByFlags(c *cli.Context, distribution bool) (*config.ArtifactoryDetails, error) {
	artDetails, err := createArtifactoryDetails(c, true)
	if err != nil {
		return nil, err
	}
	if distribution {
		if artDetails.DistributionUrl == "" {
			return nil, errors.New("the --dist-url option is mandatory")
		}
	} else {
		if artDetails.Url == "" {
			return nil, errors.New("the --url option is mandatory")
		}
	}

	return artDetails, nil
}

func validateServerId(serverId string) error {
	reservedIds := []string{"delete", "use", "show", "clear"}
	for _, reservedId := range reservedIds {
		if serverId == reservedId {
			return errors.New(fmt.Sprintf("Server can't have one of the following ID's: %s\n %s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage()))
		}
	}
	return nil
}

func configCmd(c *cli.Context) error {
	if len(c.Args()) > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var serverId string
	configCommandConfiguration, err := createConfigCommandConfiguration(c)
	if err != nil {
		return err
	}
	if len(c.Args()) == 2 {
		if c.Args()[0] == "import" {
			return commands.Import(c.Args()[1])
		}
		serverId = c.Args()[1]
		if err := validateServerId(serverId); err != nil {
			return err
		}
		artDetails, err := config.GetArtifactorySpecificConfig(serverId)
		if err != nil {
			return err
		}
		if artDetails.IsEmpty() {
			log.Info("\"" + serverId + "\" configuration could not be found.")
			return nil
		}
		if c.Args()[0] == "delete" {
			if configCommandConfiguration.Interactive {
				if !cliutils.InteractiveConfirm("Are you sure you want to delete \"" + serverId + "\" configuration?") {
					return nil
				}
			}
			return commands.DeleteConfig(serverId)
		}
		if c.Args()[0] == "export" {
			return commands.Export(serverId)
		}
	}
	if len(c.Args()) > 0 {
		if c.Args()[0] == "show" {
			return commands.ShowConfig(serverId)
		}
		if c.Args()[0] == "clear" {
			commands.ClearConfig(configCommandConfiguration.Interactive)
			return nil
		}
		serverId = c.Args()[0]
		err = validateServerId(serverId)
		if err != nil {
			return err
		}
	}
	err = validateConfigFlags(configCommandConfiguration)
	if err != nil {
		return err
	}
	configCmd := commands.NewConfigCommand().SetDetails(configCommandConfiguration.ArtDetails).SetInteractive(configCommandConfiguration.Interactive).SetServerId(serverId).SetEncPassword(configCommandConfiguration.EncPassword)
	return configCmd.Config()
}

func useCmd(c *cli.Context) error {
	var serverId string
	var err error = nil
	if len(c.Args()) == 1 {
		serverId = c.Args()[0]
		err = validateServerId(serverId)
		if err != nil {
			return err
		}
		err = commands.Use(serverId)
		if err != nil {
			return err
		}
	} else {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	return err
}

func releaseBundleTranslateChartCmd(c *cli.Context) error {
	if !(c.NArg() == 2 && c.IsSet("chart-path") && c.IsSet("docker-repo")) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	params, err := createReleaseBundleCreateUpdateParams(c, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	translateChartCmd := distribution.NewTranslateChartCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	translateChartCmd.SetRtDetails(rtDetails).SetReleaseBundleCreateParams(params).SetSourceChartPath(c.String("chart-path")).SetDockerRepo(c.String("docker-repo")).SetDryRun(c.Bool("dry-run"))
	return commands.Exec(translateChartCmd)
}

func offerConfig(c *cli.Context) (*config.ArtifactoryDetails, error) {
	var exists bool
	exists, err := config.IsArtifactoryConfExists()
	if err != nil || exists {
		return nil, err
	}

	var ci bool
	if ci, err = clientutils.GetBoolEnvValue(cliutils.CI, false); err != nil {
		return nil, err
	}
	var offerConfig bool
	if offerConfig, err = clientutils.GetBoolEnvValue(cliutils.OfferConfig, !ci); err != nil {
		return nil, err
	}
	if !offerConfig {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return nil, nil
	}

	msg := fmt.Sprintf("To avoid this message in the future, set the %s environment variable to false.\n"+
		"The release bundle generator commands require the Artifactory URL and authentication details\n"+
		"Configuring the release bundle generator with these parameters now will save you having to include them as command options.\n"+
		"You can also configure these parameters later using the 'release-bundle-generator c' command.\n"+
		"Configure now?", cliutils.OfferConfig)
	confirmed := cliutils.InteractiveConfirm(msg)
	if !confirmed {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return nil, nil
	}
	details, err := createArtifactoryDetails(c, false)
	if err != nil {
		return nil, err
	}
	encPassword := c.BoolT("enc-password")
	configCmd := commands.NewConfigCommand().SetDefaultDetails(details).SetInteractive(true).SetEncPassword(encPassword)
	err = configCmd.Config()
	if err != nil {
		return nil, err
	}

	return configCmd.RtDetails()
}

func createConfigCommandConfiguration(c *cli.Context) (configCommandConfiguration *commands.ConfigCommandConfiguration, err error) {
	configCommandConfiguration = new(commands.ConfigCommandConfiguration)
	configCommandConfiguration.ArtDetails, err = createArtifactoryDetails(c, false)
	if err != nil {
		return
	}
	configCommandConfiguration.EncPassword = c.BoolT("enc-password")
	configCommandConfiguration.Interactive = cliutils.GetInteractiveValue(c)
	return
}

func validateConfigFlags(configCommandConfiguration *commands.ConfigCommandConfiguration) error {
	if !configCommandConfiguration.Interactive && configCommandConfiguration.ArtDetails.Url == "" {
		return errors.New("the --url option is mandatory when the --interactive option is set to false or the CI environment variable is set to true.")
	}
	return nil
}

func createArtifactoryDetails(c *cli.Context, includeConfig bool) (details *config.ArtifactoryDetails, err error) {
	if includeConfig {
		details, err := offerConfig(c)
		if err != nil {
			return nil, err
		}
		if details != nil {
			return details, err
		}
	}
	details = new(config.ArtifactoryDetails)
	details.Url = c.String("url")
	details.DistributionUrl = c.String("dist-url")
	details.ApiKey = c.String("apikey")
	details.User = c.String("user")
	details.Password = c.String("password")
	details.SshKeyPath = c.String("ssh-key-path")
	details.SshPassphrase = c.String("ssh-passphrase")
	details.AccessToken = c.String("access-token")
	details.ClientCertPath = c.String("client-cert-path")
	details.ClientCertKeyPath = c.String("client-cert-key-path")
	details.ServerId = c.String("server-id")
	details.InsecureTls = c.Bool("insecure-tls")
	if details.ApiKey != "" && details.User != "" && details.Password == "" {
		// The API Key is deprecated, use password option instead.
		details.Password = details.ApiKey
		details.ApiKey = ""
	}

	if includeConfig && !credentialsChanged(details) {
		confDetails, err := commands.GetConfig(details.ServerId)
		if err != nil {
			return nil, err
		}

		if details.Url == "" {
			details.Url = confDetails.Url
		}
		if details.DistributionUrl == "" {
			details.DistributionUrl = confDetails.DistributionUrl
		}

		if !isAuthMethodSet(details) {
			if details.ApiKey == "" {
				details.ApiKey = confDetails.ApiKey
			}
			if details.User == "" {
				details.User = confDetails.User
			}
			if details.Password == "" {
				details.Password = confDetails.Password
			}
			if details.SshKeyPath == "" {
				details.SshKeyPath = confDetails.SshKeyPath
			}
			if details.AccessToken == "" {
				details.AccessToken = confDetails.AccessToken
			}
			if details.RefreshToken == "" {
				details.RefreshToken = confDetails.RefreshToken
			}
			if details.TokenRefreshInterval == cliutils.TokenRefreshDisabled {
				details.TokenRefreshInterval = confDetails.TokenRefreshInterval
			}
			if details.ClientCertPath == "" {
				details.ClientCertPath = confDetails.ClientCertPath
			}
			if details.ClientCertKeyPath == "" {
				details.ClientCertKeyPath = confDetails.ClientCertKeyPath
			}
		}
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	details.DistributionUrl = clientutils.AddTrailingSlashIfNeeded(details.DistributionUrl)

	err = config.CreateInitialRefreshTokensIfNeeded(details)
	return
}

func credentialsChanged(details *config.ArtifactoryDetails) bool {
	return details.Url != "" || details.User != "" || details.Password != "" ||
		details.ApiKey != "" || details.SshKeyPath != "" || details.SshAuthHeaderSet() ||
		details.AccessToken != ""
}

func isAuthMethodSet(details *config.ArtifactoryDetails) bool {
	return (details.User != "" && details.Password != "") || details.SshKeyPath != "" || details.ApiKey != "" || details.AccessToken != ""
}

func createReleaseBundleCreateUpdateParams(c *cli.Context, bundleName, bundleVersion string) (distributionServicesUtils.ReleaseBundleParams, error) {
	releaseBundleParams := distributionServicesUtils.NewReleaseBundleParams(bundleName, bundleVersion)
	releaseBundleParams.SignImmediately = c.Bool("sign")
	releaseBundleParams.StoringRepository = c.String("repo")
	releaseBundleParams.GpgPassphrase = c.String("passphrase")
	releaseBundleParams.Description = c.String("desc")
	if c.IsSet("release-notes-path") {
		bytes, err := ioutil.ReadFile(c.String("release-notes-path"))
		if err != nil {
			return releaseBundleParams, errorutils.CheckError(err)
		}
		releaseBundleParams.ReleaseNotes = string(bytes)
		releaseBundleParams.ReleaseNotesSyntax, err = populateReleaseNotesSyntax(c)
		if err != nil {
			return releaseBundleParams, err
		}
	}
	return releaseBundleParams, nil
}

func populateReleaseNotesSyntax(c *cli.Context) (distributionServicesUtils.ReleaseNotesSyntax, error) {
	// If release notes syntax is set, use it
	releaseNotexSyntax := c.String("release-notes-syntax")
	if releaseNotexSyntax != "" {
		switch releaseNotexSyntax {
		case "markdown":
			return distributionServicesUtils.Markdown, nil
		case "asciidoc":
			return distributionServicesUtils.Asciidoc, nil
		case "plain_text":
			return distributionServicesUtils.PlainText, nil
		default:
			return distributionServicesUtils.PlainText, errorutils.CheckError(errors.New("--release-notes-syntax must be one of: markdown, asciidoc or plain_text."))
		}
	}
	// If the file extension is ".md" or ".markdown", use the markdonwn syntax
	extension := strings.ToLower(filepath.Ext(c.String("release-notes-path")))
	if extension == ".md" || extension == ".markdown" {
		return distributionServicesUtils.Markdown, nil
	}
	return distributionServicesUtils.PlainText, nil
}
