package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Désinstalle HiberCli",
	Run: func(cmd *cobra.Command, args []string) {
		vp := viper.New()
		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(configDir)
		err := vp.ReadInConfig()
		if err != nil {
			red.Println(err)
			os.Exit(0)
		}
		//Définir hibercliSize comme une variable globale
		var hibercliSize int64
		path, _ := exec.LookPath(vp.GetString("cli.command"))
		//Vérifier la taille du fichier de configuration
		config, err := os.Stat(configDir)
		//Si il y a une erreur ne pas le calculer
		if err != nil {
			hibercliSize = 0
		} else {
			hibercliSize = config.Size()
		}
		//Vérifier la taille du dossier de configuration
		temp, err := os.Stat(os.TempDir() + "/HiberCLI_temp/")
		//Si il y a une erreur ne pas le calculer
		if err != nil {
			hibercliSize += 0
		} else {
			// Sinon ajouter la taille du dossier de configuration au calcul
			hibercliSize += temp.Size()
		}
		//Vérifier la taille du fichier hibercli
		hibercli, err := os.Stat(path)
		//Si il y a une erreur ne pas le calculer
		if err != nil {
			hibercliSize += 0
		} else {
			// Sinon ajouter la taille du fichier hibercli au calcul
			hibercliSize += hibercli.Size()
		}

		fmt.Println("• HiberCli ne fonctionne pas correctement ? " + bmagenta.Sprint("Ouvrez une issue !\n"))
		fmt.Println("Estimation de l'espace disque qui sera libéré : " + bmagenta.Sprintf(readableSize(hibercliSize)))
		var choice string
		inquirer = &survey.Select{
			Message: "Désinstaller HiberCLI ? ",
			Options: []string{
				bgreen.Sprintf("Non"),
				red.Sprintf("Oui"),
			},
		}
		survey.AskOne(inquirer, &choice)
		if choice == bgreen.Sprintf("Non") {
			bgreen.Println("Merci pour votre confiance !")
		}
		if choice == red.Sprintf("Oui") {
			bar := progressbar.NewOptions64(hibercliSize,
				progressbar.OptionSetDescription(bmagenta.Sprint("Suppression de HiberCLI\n")),
			)
			if _, err := os.Stat(configDir); !os.IsNotExist(err) {
				err := os.RemoveAll(configDir)
				if err != nil {
					red.Println("Erreur lors de la suppression du dossier de configuration : ", err, "PATH : ", configDir)
					os.Exit(0)
				}
				bar.Add64(config.Size())
				if _, err := os.Stat(os.TempDir() + "/HiberCLI_temp/"); !os.IsNotExist(err) {
					err := os.RemoveAll(os.TempDir() + "/HiberCLI_temp/")
					bar.Add64(int64(temp.Size()))
					if err != nil {
						red.Println("Erreur lors de la suppression du dossier temporaire :", err, "PATH : ", os.TempDir()+"/HiberCLI_temp/")
						
					}
				}
				path, err := exec.LookPath(vp.GetString("cli.command"))
				if err != nil {
					red.Println("HiberCli n'a pas été détecté sur votre système.")
					os.Exit(0)
				}
				err = os.Remove(path)
				if err != nil {
					red.Println("Erreur lors de la suppression de HiberCli :", err, "PATH : ", path, "\nEssayer de refaire la commande en tant qu'administrateur/sudoeur ou de le supprimer vous même.")
					os.Exit(0)
				}
			}
			bgreen.Println("HiberCli a été désinstallé avec succès. Merci d'avoir utiliser HiberCli !")
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
