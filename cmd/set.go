package cmd

import (
	"errors"
	"os"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/inancgumus/screen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tempDir        = os.TempDir() + "FreeTransCLI_temp"
	historicfile   = os.TempDir() + "FreeTransCLI_temp/historic.yaml"
	unzipchoice    string
	notifychoice   string
	soundchoice    string
	iconchoice     string
	clipchoice     string
	qrchoice       string
	histchoice     string
	updatechoice   string
	notfoundchoice string
	inquirer       *survey.Select
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Paramétrer le client",
	Long: `
Paramétrer et personnaliser le client, pour le moment vous pouvez changer le dossier par défaut pour les téléchargements, les couleurs et le spinner.
`,

	Run: func(cmd *cobra.Command, args []string) {
		//Conf()
		vp := viper.New()
		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(configDir)
		err := vp.ReadInConfig()
		if err != nil {
			red.Println(err)
			os.Exit(0)
		}

		for {
			if vp.GetBool("cli.unzip") {
				unzipchoice = "Désactiver la décompression automatique"
			} else {
				unzipchoice = "Activer la décompression automatique"
			}
			if vp.GetBool("cli.notify") {
				notifychoice = "Désactiver les notifications"
			} else {
				notifychoice = "Activer les notifications"
			}
			if vp.GetBool("cli.sound") {
				soundchoice = "Désactiver le son des notifications"
			} else {
				soundchoice = "Activer le son des notifications"
			}
			if runtime.GOOS != "darwin" && vp.GetBool("cli.notify") {
				iconchoice = "Choisir une icône"
			} else if runtime.GOOS == "darwin" {
				iconchoice = color.HiBlackString("Choisir une icône (indisponible sur macOS)")
			} else {
				iconchoice = color.HiBlackString("Choisir une icône (indisponible sans notification)")
			}
			if vp.GetBool("cli.clipboard") {
				clipchoice = "Désactiver le copier-coller automatique"
			} else {
				clipchoice = "Activer le copier-coller automatique"
			}
			if vp.GetBool("cli.qrcode") {
				qrchoice = "Désactiver l'affichage du QR code"
			} else {
				qrchoice = "Activer l'affichage du QR code"
			}
			if vp.GetBool("cli.history") {
				histchoice = "Désactiver l'historique"
			} else {
				histchoice = "Activer l'historique"
			}
			if vp.GetBool("cli.update") {
				updatechoice = "Désactiver l'avertissement de mise à jour"
			} else {
				updatechoice = "Activer l'avertissement de mise à jour"
			}

			if vp.GetBool("cli.notfound") {
				notfoundchoice = "Désactiver la suggestion de chemin similaire en cas d'erreur"
			} else {
				notfoundchoice = "Activer la suggestion de chemin similaire en cas d'erreur"
			}

			var choice string
			if os.Getenv("DISPLAY") != "" || os.Getenv("DISPLAY") != ":0" || runtime.GOOS == "windows" {
				inquirer = &survey.Select{
					Message: "Quel paramètre ? " + color.HiBlackString("(CTRL+C pour quitter)"),
					Options: []string{
						"Choisir le dossier de téléchargement par défaut",
						unzipchoice,
						notifychoice,
						soundchoice,
						iconchoice,
						clipchoice,
						qrchoice,
						histchoice,
						updatechoice,
						notfoundchoice,
						red.Sprint("Effacer l'historique"),
						red.Sprint("Réinitialiser la configuration"),
						red.Sprint("Désinstaller FreeTransCLI"),
					},
					PageSize: 13,
				}
			} else {
				inquirer = &survey.Select{
					Message: "Quel paramètre ?" + color.HiBlackString("(CTRL+C pour quitter)a"),
					Options: []string{
						"Choisir le dossier de téléchargement par défaut",
						unzipchoice,
						clipchoice,
						qrchoice,
						histchoice,
						updatechoice,
						notfoundchoice,
						red.Sprint("Effacer l'historique"),
						red.Sprint("Réinitialiser la configuration"),
						red.Sprint("Désinstaller FreeTransCLI"),
					},
					PageSize: 10,
				}
			}
			err := survey.AskOne(inquirer, &choice)
			if err != nil {
				if err == terminal.InterruptErr {
					screen.Clear()
					screen.MoveTopLeft()
					os.Exit(0)
				}
			}
			if choice == "Choisir le dossier de téléchargement par défaut" {
				var dldpath string
				prompt := &survey.Input{
					Message: "Chemin du dossier de téléchargement :",
					Suggest: func(toComplete string) []string {
						return []string{home + "/Downloads", "./"}
					}}
				survey.AskOne(prompt, &dldpath)
				//enlever les ' au début et à la fin
				dldpath = strings.Trim(dldpath, "'")

				if len(dldpath) == 0 {
					red.Println("Erreur : Aucun chemin spécifié")
					continue
				}
				dir, err := os.Stat(dldpath)
				if errors.Is(err, os.ErrNotExist) {
					red.Println("Erreur : Dossier introuvable")
					continue
				}
				if !dir.IsDir() {
					red.Println("Erreur : Le chemin spécifié n'est pas un dossier")
					continue
				}

				//vérifier si le dossier est accessible en écriture
				f, err := os.Create(dldpath + "/test.txt")
				if err != nil {
					red.Println("Erreur : Le dossier spécifié n'est pas accessible en écriture")
					continue
				}
				f.Close()
				os.Remove(dldpath + "/test.txt")

				vp.Set("cli.dld", dldpath)

				green.Println("Le dossier de téléchargement par défaut a été changé avec succès !")
			}
			// if choice == "Choisir un spinner" {
			// 	spinchoice := true
			// 	prompt := &survey.Confirm{
			// 		Message: "Pour choisir un spinner le programme doit ouvrir votre navigateur web. (n pour afficher un QRcode)",
			// 		Default: true,
			// 	}
			// 	survey.AskOne(prompt, &spinchoice)
			// 	if spinchoice {
			// 		openbrowser("https://github.com/briandowns/spinner#available-character-sets")
			// 	} else {
			// 		qrterminal.GenerateHalfBlock(("https://github.com/briandowns/spinner#available-character-sets"), qrterminal.M, os.Stdout)
			// 		fmt.Print("https://github.com/briandowns/spinner#available-character-sets\n\n")

			// 	}
			// 	cyan.Print("Quel est l'index du spinner que vous voulez utiliser ? (0 à 43) ")
			// 	var spin string
			// 	fmt.Scanln(&spin)
			// 	//Vérifier si le spinner est vide
			// 	if len(spin) == 0 {
			// 		red.Println("Erreur : Vous devez spécifier un spinner")
			// 		continue
			// 	}
			// 	//convertir le string en int
			// 	spinint, err := strconv.Atoi(spin)

			// 	if err != nil {
			// 		red.Println("Erreur : Vous devez spécifier un nombre")
			// 		continue
			// 	}
			// 	//vérifier si le spinner est entre 0 et 43
			// 	if spinint < 0 || spinint > 43 {
			// 		red.Println("Erreur : Vous devez spécifier un nombre entre 0 et 43")
			// 		continue
			// 	}
			// 	//set le spinner dans le fichier de config
			// 	vp.Set("cli.spinner", spinint)
			// 	s := spinner.New(spinner.CharSets[spinint], 100*time.Millisecond)
			// 	s.Prefix = color.GreenString("Spinner choisi : ")
			// 	s.Start()
			// 	time.Sleep(2 * time.Second)
			// 	s.Stop()
			// }

			if choice == unzipchoice {
				vp.Set("cli.unzip", !vp.GetBool("cli.unzip"))
			}
			if choice == notifychoice {
				vp.Set("cli.notify", !vp.GetBool("cli.notify"))
			}

			if choice == soundchoice {
				vp.Set("cli.sound", !vp.GetBool("cli.sound"))
			}
			if choice == iconchoice {
				if runtime.GOOS == "darwin" || !vp.GetBool("cli.notify") {
					continue
				}
				var iconpath string
				prompt := &survey.Input{
					Message: "Chemin de l'icône :",
				}
				survey.AskOne(prompt, &iconpath)
				//enlever les ' au début et à la fin
				iconpath = strings.Trim(iconpath, "'")
				if len(iconpath) == 0 {
					continue
				}
				file, err := os.Stat(iconpath)
				if errors.Is(err, os.ErrNotExist) {
					red.Println("Erreur : Fichier introuvable")
					continue
				}
				// Vérifier que le fichier est une image
				if !strings.Contains(file.Name(), ".png") && !strings.Contains(file.Name(), ".jpg") && !strings.Contains(file.Name(), ".jpeg") {
					red.Println("Erreur : Le fichier spécifié n'est pas une image")
					continue
				}
				var selecticon string
				if vp.GetBool("cli.sound") {
					beeep.Alert("FreeTransCLI", "Test des notifications", iconpath)
				} else if !vp.GetBool("cli.sound") {
					beeep.Notify("FreeTransCLI", "Test des notifications", iconpath)
				}
				notifinquirer := &survey.Select{
					Message: "Quel paramètre ? (CTRL+C pour quitter)",
					Options: []string{
						"Garder cette icône",
						"Revenir sur l'ancienne icône",
						"Revenir sur l'icône par défaut",
						"Changer l'icône",
					}}
				survey.AskOne(notifinquirer, &selecticon)
				if selecticon == "Garder cette icône" {
					//Déplacer l'icône dans le dossier de config
					os.Rename(iconpath, configDir+"/icon.png")
					vp.Set("cli.icon", iconpath)
					green.Println("L'icône a été gardée")
					continue
				}
				if selecticon == "Revenir sur l'ancienne icône" {
					green.Println("L'icône n'a pas été changée")
					continue
				}
				if selecticon == "Revenir sur l'icône par défaut" {
					vp.Set("cli.icon", "")
					green.Println("L'icône a été réinitialisée")
					continue
				}
				if selecticon == "Changer l'icône" {
					continue
				}
			}

			if choice == clipchoice {
				vp.Set("cli.clipboard", !vp.GetBool("cli.clipboard"))
			}

			if choice == qrchoice {
				vp.Set("cli.qrcode", !vp.GetBool("cli.qrcode"))

			}
			if choice == histchoice {
				vp.Set("cli.history", !vp.GetBool("cli.history"))
			}

			if choice == updatechoice {
				vp.Set("cli.update", !vp.GetBool("cli.update"))
			}

			if choice == notfoundchoice {
				vp.Set("cli.notfound", !vp.GetBool("cli.notfound"))
			}

			if choice == red.Sprint("Effacer l'historique") {
				file, err := os.Stat(historicfile)
				if errors.Is(err, os.ErrNotExist) {
					red.Println("Erreur : Fichier introuvable, il sera recréé automatiquement lors d'un upload")
					continue
				}
				delchoice := false
				prompt := &survey.Confirm{
					Message: bred.Sprint("\rSouhaitez-vous effacer l'historique ?\n") + readableSize(file.Size()) + " seront libérés",
				}
				survey.AskOne(prompt, &delchoice)
				if delchoice {
					err := os.Remove(historicfile)
					if err != nil {
						red.Println("Erreur : Impossible d'effacer l'historique")
						continue
					}
					green.Println("L'historique a été effacé")
					continue
				} else {
					yellow.Println("L'historique n'a pas été effacé")
					continue
				}

			}
			if choice == red.Sprint("Réinitialiser la configuration") {
				reset := false
				prompt := &survey.Confirm{
					Message: bred.Sprint("Souhaitez vous réinitialiser la configuration ? Attention : Cette action est irréversible !"),
				}
				survey.AskOne(prompt, &reset)
				if reset {
					err := os.Remove(configFilePath)
					if err != nil {
						red.Println("Erreur : Impossible de réinitialiser la configuration")
						continue
					}
					green.Println("La configuration a été réinitialisée")
					continue
				} else {
					yellow.Println("La configuration n'a pas été réinitialisée")
					continue
				}
			}
			if choice == red.Sprint("Désinstaller FreeTransCLI") {
				uninstallCmd.Run(uninstallCmd, []string{})
			}
			err = vp.WriteConfig()
			if err != nil {
				red.Println("Erreur : Impossible d'écrire la configuration\n", err)
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Aliases = []string{"setting", "settings", "config", "conf", "s", "c"}
	setCmd.DisableFlagsInUseLine = true

	setCmd.SetHelpTemplate(`{{.Long}}

Usage:
      freetranscli set

Aliases:
	set, setting, settings, config, conf
`)
}
