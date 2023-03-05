package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bgreen         = color.New(color.Bold, color.FgGreen)
	red            = color.New(color.FgRed)
	bred           = color.New(color.Bold, color.FgRed)
	bmagenta       = color.New(color.Bold, color.FgMagenta)
	yellow         = color.New(color.FgYellow)
	green          = color.New(color.FgGreen)
	cyan           = color.New(color.FgCyan)
	configFilePath string
	configDir      string
	dldPath        string
	home           string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "freetranscli",
}

func Conf() {
	//Définir le répertoire de configuration selon l'OS
	if runtime.GOOS == "windows" {
		configDir = os.Getenv("APPDATA") + "/freetranscli" //C:\Users\%USERNAME%\AppData\Roaming\freetranscli
		home = os.Getenv("USERPROFILE")                    //C:\Users\%USERNAME%

	} else {
		configDir = os.Getenv("HOME") + "/.config/freetranscli" //~/.config/freetranscli
		home = os.Getenv("HOME")                                //~

	}
	// créer le répertoire de configuration s'il n'existe pas
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, 0777)
	}

	// créer le fichier de configuration s'il n'existe pas
	configFilePath = configDir + "/config.yaml"
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		os.Create(configFilePath)
	}

	//Si il n'y a pas de dossier temporaire, on en crée un
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0777)
	}

	//Créer un fichier historic.yaml
	if _, err := os.Stat(tempDir + "/historic.yaml"); os.IsNotExist(err) {
		os.Create(tempDir + "/historic.yaml")
	}
	vp := viper.New()

	//Définir le chemin de téléchargement par défaut
	_, err := os.Stat(home + "/Downloads") //Vérifier si le dossier Downloads existe
	if err != nil {
		dldPath = "./" //Si il n'existe pas on télécharge dans le dossier où se trouve l'utilisateur
	} else {
		dldPath = home + "/Downloads" //Sinon on télécharge dans le dossier Downloads
	}
	// Initialise la configuration
	config := map[string]interface{}{
		"cli.clipboard": true,
		"cli.dld":       dldPath,
		"cli.notify":    true,
		"cli.icon":      "",
		"cli.sound":     true,
		"cli.spinner":   14,
		"cli.qrcode":    true,
		"cli.history":   true,
		"cli.update":    true,
		"cli.lastmsg":   "",
		"cli.notfound":  true,
		"cli.unzip":     true,
	}

	// Lit la configuration existante
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath(configDir)
	err = vp.ReadInConfig()
	if err != nil {
		red.Println(err)
		os.Exit(0)
	}

	// Vérifie si toutes les clés de configuration existent et ajoute les valeurs par défaut si nécessaire
	for key, value := range config {
		if !vp.IsSet(key) {
			vp.Set(key, value)
		}
	}

	// Écrit la configuration
	err = vp.WriteConfig()
	if err != nil {
		red.Println("Impossible d'écrire la configuration :", err)
		os.Exit(0)
	}

	var (
		currentDate    = time.Now()
		difference     = currentDate.Sub(vp.GetTime("cli.lastmsg"))
		currentVersion = "0.0.0"
	)
	//Obtenir la dernière version publiée
	resp, err := http.Get("https://api.github.com/repos/el2zay/freetranscli/releases/latest")
	if err != nil {
		yellow.Println("Impossible de récupérer la dernière version publiée\n", err)
	}

	//Lire le body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		red.Println(err)
		os.Exit(0)
	}

	//Convertir en json
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	//Vérifier que si la version actuelle est inférieure à la dernière version publiée et que l'utilisateur a activé l'affichage du message et que le dernier message a été affiché il y a plus de 12 heures
	if currentVersion < data["tag_name"].(string) && vp.GetBool("cli.update") && difference.Hours() >= 12 {
		fmt.Print("Une nouvelle version est disponible ", bred.Sprint(currentVersion), " → ", bgreen.Sprint(data["tag_name"]), "\n'freetranscli set' pour activer les mises à jour automatiques \n\n")

		vp.Set("cli.lastmsg", currentDate)
	}
	//Ecrire dans la configuration
	err = vp.WriteConfig()
	if err != nil {
		red.Println("Impossible d'écrire la configuration\n", err)
		os.Exit(0)
	}

}

// Tout le temps executer au démarrage
func Execute() {
	Conf()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpTemplate(`
Usage:
      {{.Use}} [commande]

Commandes:
      download/d    Télécharger un fichier depuis FreeTransfert grâce à l'url du fichier
      help          Aide à propos d'une commande
      history       Affiche l'historique des fichiers téléversés
      issue         Ouvre une issue sur GitHub
      set/config    Paramétrer FreeTransCLI
      uninstall     Désinstaller FreeTransCLI
      upload/u      Téléverser un fichier sur FreeTransfert grâce au chemin du fichier
`)

}
