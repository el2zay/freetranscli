package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Affiche l'historique des fichiers téléversés",
	Long:  `Affiche l'historique des fichiers téléversés avec la date, le chemin du fichier au moment du téléversement et l'url HiberFile`,
	Run: func(cmd *cobra.Command, args []string) {

		//Executer la commande open os.TempDir() + "/HiberCLI_temp/historic.yaml
		err := exec.Command("open", os.TempDir()+"/HiberCLI_temp/historic.yaml").Run()

		if err != nil {
			fmt.Println(err)
		}
	},
}

func historic(url string, path string, filetype string, size string) {
	//Afficher la date et l'heure au format DD/MM/YYYY HH:MM:SS
	vp := viper.New()
	vp.SetConfigName("historic")
	vp.SetConfigType("yaml")
	vp.AddConfigPath(os.TempDir() + "/HiberCLI_temp/")
	err := vp.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	now := time.Now()
	dateTimeString := now.Format("02/01/2006 15:04:05")

	vp.Set(dateTimeString+".path", path)
	vp.Set(dateTimeString+".url", url)
	vp.Set(dateTimeString+".filetype", filetype)
	vp.Set(dateTimeString+".size", size)

	err = vp.WriteConfig()
	if err != nil {
		red.Println("Erreur : Impossible d'écrire la configuration\n", err)
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(historyCmd)

}
