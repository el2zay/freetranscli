package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mdp/qrterminal"
	"github.com/spf13/cobra"
)

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux", "android":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		qrterminal.GenerateHalfBlock((url), qrterminal.M, os.Stdout)
		red.Println("Erreur lors de l'ouverture du navigateur", url)
	}
}

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Ouvre une issue sur GitHub",
	Long: `
Ouvre sur votre navigateur web la page des issues sur GitHub pour signaler un bug ou demander une fonctionnalité.
Si vous n'avez pas de navigateur web l'url s'affichera dans la console.
Si vous n'avez pas de compte GitHub vous pouvez en créé un gratuitement sur https://github.com/signup`,

	Run: func(cmd *cobra.Command, args []string) {
		//Demander le titre
		var title string
		prompt := &survey.Input{
			Message: "Titre de l'issue:",
		}
		survey.AskOne(prompt, &title)

		//Demander la description avec survey en multi ligne
		var description string
		multiline := &survey.Multiline{
			Message: "Description de l'issue:",
		}
		survey.AskOne(multiline, &description)

		//Ouvrir le site d'issue avec le titre et la description
		openbrowser("https://github.com/el2zay/hibercli/issues/new?title=" + title + "&body=" + description)
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
	issueCmd.Aliases = []string{"bug", "github"}
}
