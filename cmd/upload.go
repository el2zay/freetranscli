package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/atotto/clipboard"
	"github.com/gen2brain/beeep"
	"github.com/mdp/qrterminal"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	filetype = "file"
	url      string
	i        int
	size     int64

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
	//Créer une progressbar qui affiche la taille totale des fichiers à archiver
	bar := progressbar.DefaultBytes(size, cyan.Sprint("Archivage"))
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
			//Mettre à jour la progressbar
			bar.Add64(info.Size())
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

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Téléverser un fichier sur FreeTransCLI grâce au chemin du fichier",
	Long: `
Téléverser un fichier sur FreeTransCLI grâce au chemin du fichier sur votre ordinateur.
Exemple : freetranscli upload /Users/username/Documents/Hey.mov

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
			args[i] = strings.ReplaceAll(args[i], "'", "")
		}

		//Vérifier qu'il n y a aucune erreur dans les fichiers
		for i := len(args) - 1; i >= 0; i-- {
			//Retirer le / a la fin du chemin si il y en a un
			if args[i][len(args[i])-1:] == "/" {
				args[i] = args[i][:len(args[i])-1]
			}
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
				continue
			}

			//Si le fichier n'a pas les droits d'accès, on affiche une erreur
			if os.IsPermission(err) {
				red.Println("Erreur : Permission refusée, vérifiez que vous avez les droits d'accès au fichier, avez-vous lancé le programme en tant qu'administrateur/sudoeur ?")
				continue
			}

			file, _ := os.Stat(args[i])
			size := file.Size()

			//si le fichier est plus gros que 50go, on affiche une erreur
			if size > 50000000000 {
				red.Println("Erreur : Vous ne pouvez pas upload un fichier plus gros que 50Go.")
				os.Exit(0)
			}

			//si dans args il y a plus d'un fichier on attend de récupérer tous les fichiers pour les mettres dans un dossier
			if len(args) > 1 {
				// Si le dossier temporaire n'existe pas alors on le créé
				if _, err := os.Stat(tempDir + "/free-transfert"); os.IsNotExist(err) {
					err := os.Mkdir(tempDir+"/free-transfert", 0755)
					if err != nil {
						red.Println(err)
						os.Exit(0)
					}
				}
				//Executer la commande cp
				err := exec.Command("cp", "-R" , args[i], tempDir+"/free-transfert").Run()

				if err != nil {
					red.Println(err)
					os.Exit(0)
				}
				if i > 0 {
					//Revenir au début de la boucle
					continue
				}
				if i == 0 {
					args[i] = tempDir + "/free-transfert" + ".zip"

					//Archiver le dossier
					if err := zipSource(tempDir+"/free-transfert", args[i]); err != nil {
						red.Println(err)
						os.Exit(0)
					}
					// Supprimer le dossier temporaire
					err := os.RemoveAll(tempDir + "/free-transfert")
					if err != nil {
						red.Println(err)
						os.Exit(0)
					}
				}
				file, _ := os.Stat(args[i])
				size = file.Size()
			}
			//Si c'est un dossier
			if file.IsDir() && len(args) == 1 {
				//si le dossier est plus gros que 50go
				if size > 50000000000 {
					red.Println("Erreur : Vous ne pouvez pas upload un dossier plus gros que 50Go.")
					os.Exit(0)
				}
				//Archiver le dossier
				if err := zipSource(args[0], args[i]+".zip"); err != nil {
					red.Println(err, "Désolé essayez de le compresser vous même…")
					//Supprimer définitivement le fichier
					os.Remove(args[i] + ".zip")
					os.Exit(0)
				} else if err == nil {
					//Supprimer définitivement le fichier
					args[i] = args[i] + ".zip"
					file, _ := os.Stat(args[i])
					size = file.Size()
					// os.Remove(args[i] + ".zip")
				}
			}
			//Progress bar pour le téléversement
			bar := progressbar.DefaultBytes(
				size,
				green.Sprint("Téléversement"),
			)
			//Ajouter un fichier au reader
			reader, err := os.Open(args[i])
			if err != nil {
				red.Println(err)
				os.Exit(0)
			}

			//Faire avancer la progressbar
			_, err = io.Copy(bar, reader)
			if err != nil {
				red.Println(err)
				os.Exit(0)
			}
			//Supprimer la progressbar
			// fmt.Print("\033[2K\r")

		} //Fin de la boucle for
		url = "https://github.com/mdp/qrterminal"

		//Enregistre les données dans un fichier d'historique si l'historique est activé
		if vp.GetBool("cli.history") {
			absPath, _ := filepath.Abs(args[i])                  //Chemin des fichiers
			historic(url, absPath, filetype, readableSize(size)) //Enregistrer dans l'historique avec l'url, le chemin du fichier, le type de fichier et la taille du fichier
		}

		//Vérifier si il faut afficher une notification et si il le faut avec du son.
		if vp.GetBool("cli.notify") && vp.GetBool("cli.sound") {
			beeep.Alert("FreeTransCLI", "Votre fichier a bien été upload.", vp.GetString("cli.icon"))
		} else if vp.GetBool("cli.notify") && !vp.GetBool("cli.sound") {
			beeep.Notify("FreeTransCLI", "Votre fichier a bien été upload.", vp.GetString("cli.icon"))
		}

		//Vérifier si il faut afficher le qrcode
		if vp.GetBool("cli.qrcode") {
			//Imprimer le qrcode en petit
			qrterminal.GenerateHalfBlock((url), qrterminal.L, os.Stdout)
		}
		//Vérifier si il faut copier l'adresse dans le presse-papier
		if vp.GetBool("cli.clipboard") {
			clipboard := clipboard.WriteAll(url)
			if clipboard != nil {
				yellow.Println("Scannez le QR code pour télécharger votre fichier, l'adresse n'a pas pu être copiée dans votre presse-papiers.", url)
				os.Exit(0)
			}
			green.Println("Scannez le QR code pour télécharger votre fichier, l'adresse est copiée dans votre presse-papiers.")
		} else if vp.GetBool("cli.qrcode") {
			green.Println("\rScannez le QR code pour télécharger votre fichier.", url)
		} else {
			green.Println("Votre fichier est disponible à l'adresse suivante :", url)
		}

	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.SetUsageTemplate("Usage: freetranscli upload [file]\n\n")
	uploadCmd.Aliases = []string{"up", "u", "upld"}
}
