package cmd

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Télécharger un fichier depuis HiberFile grâce à l'url du fichier",
	Long: `
Télécharger un fichier depuis HiberFile grâce à l'url du fichier
Exemple : hibercli download https://hibercli.fr/2kxQZv
Le fichier sera téléchargé dans le dossier dans lequel vous vous trouvez

Alias : d, dld, dl, down`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			var input string
			prompt := &survey.Input{
				Message: "Lien HiberCli :",
			}
			survey.AskOne(prompt, &input)
			args = append(args, input)
			println(args[0])
			//Retirer les guillemets si il y en a
			args[0] = strings.ReplaceAll(args[0], "'", "")
		}
		//code here
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.SetUsageTemplate("Usage: hibercli download [url]\n\n")
	downloadCmd.Aliases = []string{"d", "dld", "dl", "down"}

}
