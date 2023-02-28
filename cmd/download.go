package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Télécharger un fichier depuis FreeTransfert grâce à l'url du fichier",
	Long: `
Télécharger un fichier depuis FreeTransfert grâce à l'url du fichier
Exemple : freetranscli download https://freetranscli.fr/2kxQZv
Le fichier sera téléchargé dans le dossier qui est enregistré dans la configuration

Alias : d, dld, dl, down`,
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
		if len(args) == 0 {
			var input string
			prompt := &survey.Input{
				Message: "Lien FreeTransCLI :",
			}
			survey.AskOne(prompt, &input)
			args = append(args, input)
			println(args[0])
			//Retirer les guillemets si il y en a
			args[0] = strings.ReplaceAll(args[0], "'", "")
		}
		//Séparer les / pour ne garder que le code du transfert
		transfertKey := strings.Split(args[0], "/")
		// Obtenir des informations sur le transfert
		resp, err := http.Get("https://api.scw.iliad.fr/freetransfert/v2/transfers/" + transfertKey[3])
		if err != nil {
			red.Printf("Erreur (1er fetch) : %s\n", err.Error())
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			red.Printf("Erreur (1er fetch) : %s\n", err.Error())
			return
		}

		var info map[string]interface{}
		if err := json.Unmarshal(body, &info); err != nil {
			fmt.Printf("Erreur (1er fetch) : %s\n", err.Error())
			return
		}

		if info["error"] != nil || info["message"] != nil {
			errMsg, ok := info["message"].(string)
			if !ok {
				errMsg = fmt.Sprintf("%v", info["message"])
			}
			red.Printf("Erreur (1er fetch) : %s\n", errMsg)
			return
		}

		var path string
		if zip, ok := info["zip"].(map[string]interface{}); ok {
			if zipPath, ok := zip["path"].(string); ok {
				path = zipPath
				fmt.Println("ceci est un zip")
			}
		}
		if path == "" {
			if files, ok := info["files"].([]interface{}); ok && len(files) > 0 {
				if filePath, ok := files[0].(map[string]interface{})["path"].(string); ok {
					path = filePath
				}
			}
		}

		resp, err = http.Get("https://api.scw.iliad.fr/freetransfert/v2/files?transferKey=" + transfertKey[3] + "&path=" + path)
		if err != nil {
			red.Printf("Erreur (2ème fetch) : %s\n", err.Error())
			return
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			red.Printf("  (2ème fetch) : %s\n", err.Error())
			return
		}

		var url map[string]interface{}
		if err := json.Unmarshal(body, &url); err != nil {
			red.Printf("Erreur (2ème fetch) : %s\n", err.Error())
			return
		}

		if url["error"] != nil || url["message"] != nil {
			errMsg, ok := url["message"].(string)
			if !ok {
				errMsg = fmt.Sprintf("%v", url["message"])
			}
			red.Printf("Erreur (2ème fetch) : %s\n", errMsg)
			return
		}

		// Télécharger le fichier
		resp, err = http.Get(url["url"].(string))
		if err != nil {
			red.Printf("Erreur lors du téléchargement : %s\n", err.Error())
			return
		}
		defer resp.Body.Close()
		fmt.Println()
		filePath := fmt.Sprintf("%s/%s", vp.Get("cli.dld"), path)

		//Vérifier si le fichier existe déjà
		if _, err := os.Stat(filePath); err == nil {
			var choice string
			inquirer = &survey.Select{
				Message: "Le fichier existe déjà, que voulez-vous faire ?",
				Options: []string{"Renommer le fichier téléchargé", "Renommer l'ancien fichier", "Remplacer", "Annuler"},
			}
			survey.AskOne(inquirer, &choice)

			if choice == "Renommer le fichier téléchargé" {
				var input string
				prompt := &survey.Input{
					Message: "Nom du fichier :",
				}
				survey.AskOne(prompt, &input)
				filePath = fmt.Sprintf("%s/%s", vp.Get("cli.dld"), input)
			}
			if choice == "Renommer l'ancien fichier" {
				var input string
				prompt := &survey.Input{
					Message: "Nom du fichier :",
				}
				survey.AskOne(prompt, &input)
				os.Rename(filePath, fmt.Sprintf("%s/%s", vp.Get("cli.dld"), input))
			}
			if choice == "Remplacer" {
				//Yes or no
				var danger bool
				inquirer := &survey.Confirm{
					Message: "Êtes-vous sûr de vouloir remplacer le fichier ?",
				}
				survey.AskOne(inquirer, &danger)
				os.Remove(filePath)
			}
			if choice == "Annuler" {
				return
			}
		}
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			green.Sprint("Téléchargement"),
		)
		out, _ := os.Create(filePath)

		if err != nil {
			red.Printf("Erreur lors du téléchargement : %s\n", err.Error())
			return
		}

		defer out.Close()

		io.Copy(io.MultiWriter(out, bar), resp.Body)

		if err != nil {
			red.Printf("Erreur lors du téléchargement : %s\n", err.Error())
			return
		}
	}}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.SetUsageTemplate("Usage: freetranscli download [url]\n\n")
	downloadCmd.Aliases = []string{"d", "dld", "dl", "down"}

}
