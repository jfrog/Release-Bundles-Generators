package completion

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/release-bundle-generators/completion/shells/bash"
	"github.com/jfrog/release-bundle-generators/completion/shells/zsh"
	"github.com/jfrog/release-bundle-generators/docs/common"
	bash_docs "github.com/jfrog/release-bundle-generators/docs/completion/bash"
	zsh_docs "github.com/jfrog/release-bundle-generators/docs/completion/zsh"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "bash",
			Usage:        bash_docs.Description,
			HelpName:     common.CreateUsage("completion bash", bash_docs.Description, bash_docs.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				bash.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Usage:        zsh_docs.Description,
			HelpName:     common.CreateUsage("completion zsh", zsh_docs.Description, zsh_docs.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				zsh.WriteZshCompletionScript()
			},
		},
	}
}
