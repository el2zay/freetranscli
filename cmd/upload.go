package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/atotto/clipboard"
	"github.com/briandowns/spinner"
	"github.com/gen2brain/beeep"
	"github.com/mdp/qrterminal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	filetype = "file"
	size     int64
	url      string
)

func zipSource(source, target string) error {
	// Compter la taille totale des fichiers à archiver
	err := filepath.Walk(source, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Créer un nouveau fichier zip
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Créer un objet zip.Writer pour écrire
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Archiver les fichiers
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// Ouvrir le fichier à archiver
			fileToZip, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fileToZip.Close()

			// Obtenir les informations sur le fichier à archiver
			info, err := fileToZip.Stat()
			if err != nil {
				return err
			}

			// Créer un header pour le fichier à archiver
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = strings.TrimPrefix(path, source+string(filepath.Separator))

			// Ajouter le fichier à archiver au fichier zip
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, fileToZip)
			if err != nil {
				return err
			}

		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Transformer la taille en octets en une taille lisible
func readableSize(size int64) string {
	const (
		KB = 1000
		MB = 1000 * KB
		GB = 1000 * MB
	)

	switch {
	//Si la taille est inférieure à 1Ko, on affiche la taille en octets
	case size < KB:
		return fmt.Sprintf("%d o", size)
		//Si la taille est inférieure à 1Mo, on affiche la taille en Ko
	case size < MB:
		return fmt.Sprintf("%.1f Ko", float64(size)/KB)
		//Si la taille est inférieure à 1Go, on affiche la taille en Mo
	case size < GB:
		return fmt.Sprintf("%.1f Mo", float64(size)/MB)
		//Sinon on affiche la taille en Go
	default:
		return fmt.Sprintf("%.1f Go", float64(size)/GB)
	}
}

func temp() {
	// Obtenir un chemin pour l'enregistrement de fichier (temporaire)
	tempDir := os.TempDir()
	//Si il n'y a pas de dossier temporaire, on en crée un
	if _, err := os.Stat(tempDir + "/HiberCLI_temp"); os.IsNotExist(err) {
		os.Mkdir(tempDir+"/HiberCLI_temp", 0777)
	}
	//Créer un fichier historic.yaml
	if _, err := os.Stat(tempDir + "/HiberCLI_temp/historic.yaml"); os.IsNotExist(err) {
		os.Create(tempDir + "/HiberCLI_temp/historic.yaml")
	}
}

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Téléverser un fichier sur Hiberfile grâce au chemin du fichier",
	Long: `
Téléverser un fichier sur Hiberfile grâce au chemin du fichier sur votre ordinateur.
Exemple : hibercli upload /Users/username/Documents/Hey.mov

Alias : up, u , upld`,
	Run: func(cmd *cobra.Command, args []string) {
		//Obtenir le fichier de configuration
		vp := viper.New()
		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(configDir)
		err := vp.ReadInConfig()
		if err != nil {
			red.Println(err)
			os.Exit(0)
		}
		//Si aucun argument n'est donné en paramètre, on affiche une erreur
		if len(args) == 0 {
			var input string
			prompt := &survey.Input{
				Message: "Chemin du fichier à téléverser :",
			}
			survey.AskOne(prompt, &input)
			args = append(args, input)
			//Retirer les guillemets si il y en a
			args[0] = strings.ReplaceAll(args[0], "'", "")

		}
		//Vérifier qu'il n ya aucune erreur dans les fichiers
		for i := 0; i < len(args); i++ {
			_, err := os.Stat(args[i])
			//Si le fichier n'existe pas, on affiche une erreur
			if os.IsNotExist(err) {
				filename := args[i]
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					red.Printf("Erreur : Le fichier %s n'existe pas, vérifiez que vous avez bien écrit le chemin du fichier.\n", filename)

					if vp.GetBool("cli.notfound") {
						dir, file := filepath.Split(filename)
						matches, err := filepath.Glob(filepath.Join(dir, "*"+file+"*"))

						if err != nil {
							red.Printf("Erreur : Impossible de rechercher des fichiers similaires pour %s : %s\n", filename, err)
							return
						}

						if len(matches) == 0 {
							red.Printf("Aucun fichier similaire trouvé pour %s\n", filename)

							if len(args) == 1 {
								return
							}

							continue
						}

						fmt.Printf("Chemin similaire trouvé pour %s : %s\n", filename, matches[0])

						var confirm bool
						prompt := &survey.Confirm{
							Message: fmt.Sprintf("Voulez-vous utiliser le chemin similaire %s ?", matches[0]),
							Default: true,
						}
						err = survey.AskOne(prompt, &confirm)

						if err != nil {
							red.Printf("Erreur : Impossible de lire la réponse de l'utilisateur : %s\n", err)
							return
						}

						if confirm {
							args[i] = matches[0]
						} else {
							continue
						}
					} else {
						if len(args) == 1 {
							return
						}

						continue
					}
				}
			}

			//Si le fichier n'a pas les droits d'accès, on affiche une erreur
			if os.IsPermission(err) {
				red.Println("Erreur : Permission refusée, vérifiez que vous avez les droits d'accès au fichier, avez-vous lancé le programme en tant qu'administrateur/sudoeur ?")
				continue
			}

			file, _ := os.Stat(args[i])
			size = file.Size()

			//si le fichier est plus gros que 50go, on affiche une erreur
			if size > 50000000000 {
				red.Println("Erreur : Vous ne pouvez pas upload un fichier plus gros que 50Go.")
				os.Exit(0)
			}

			//si le fichier est un dossier
			if file.IsDir() {
				filetype = "directory"
				//si le dossier est plus gros que 50go
				if size >= 50000000000 {
					red.Println("Erreur : Vous ne pouvez pas upload un dossier plus gros que 50Go.")
					os.Exit(0)
				}
				//Afficher le spinner
				s := spinner.New(spinner.CharSets[vp.GetInt("cli.spinner")], 100*time.Millisecond)
				s.Prefix = cyan.Sprint("Archivage du dossier en cours  ")
				s.Start()
				//Archiver le dossier
				if err := zipSource(args[0], args[0]+".zip"); err != nil {
					red.Println(err, "Désolé essayez de le compresser vous même…")
					//Supprimer définitivement le fichier
					os.Remove(args[0] + ".zip")
					os.Exit(0)
					s.Stop()
				} else {
					//Supprimer définitivement le fichier
					os.Remove(args[0] + ".zip")
					s.Stop()
				}
			}
			s := spinner.New(spinner.CharSets[vp.GetInt("cli.spinner")], 100*time.Millisecond)
			s.Prefix = green.Sprint("Envoie en cours de " + file.Name() + "  ")
			s.Start()
			time.Sleep(2 * time.Second)
			s.Stop()

			url = "https://github.com/mdp/qrterminal"

			//Enregistre les données dans un fichier d'historique si l'historique est activé
			if vp.GetBool("cli.history") {
				absPath, _ := filepath.Abs(args[i])                  //Chemin des fichiers
				temp()                                               //Executer la fonction temp pour éviter les erreurs
				historic(url, absPath, filetype, readableSize(size)) //Enregistrer dans l'historique avec l'url, le chemin du fichier, le type de fichier et la taille du fichier
			}
		} //Fin de la boucle for

		//Vérifier si il faut afficher une notification et si il le faut avec du son.
		if vp.GetBool("cli.notify") && vp.GetBool("cli.sound") {
			beeep.Alert("Hibercli", "Votre fichier a bien été upload.", vp.GetString("cli.icon"))
		} else if vp.GetBool("cli.notify") && !vp.GetBool("cli.sound") {
			beeep.Notify("Hibercli", "Votre fichier a bien été upload.", vp.GetString("cli.icon"))
		}

		//Vérifier si il faut afficher le qrcode
		if vp.GetBool("cli.qrcode") {
			qrterminal.GenerateHalfBlock((url), qrterminal.M, os.Stdout)
		}
		//Vérifier si il faut copier l'adresse dans le presse-papier
		if vp.GetBool("cli.clipboard") {
			clipboard := clipboard.WriteAll(url)
			if clipboard != nil {
				yellow.Println("Scannez le QR code pour télécharger votre fichier, l'adresse n'a pas pu être copié dans votre presse-papiers.", url)
				os.Exit(0)
			}
			green.Println("Scannez le QR code pour télécharger votre fichier, l'adresse est copié dans votre presse-papiers.")
		} else if vp.GetBool("cli.qrcode") {
			green.Println("\rScannez le QR code pour télécharger votre fichier.", url)
		} else {
			green.Println("Votre fichier est disponible à l'adresse suivante :", url)
		}

	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.SetUsageTemplate("Usage: hibercli upload [file]\n\n")
	uploadCmd.Aliases = []string{"up", "u", "upld"}
}
