package cmd

import (
	"encoding/json"
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
	magenta        = color.New(color.FgMagenta)
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
	Use: "hibercli",
}

func Conf() {
	//Définir le répertoire de configuration selon l'OS
	if runtime.GOOS == "windows" {
		configDir = os.Getenv("APPDATA") + "/hibercli"
		home = os.Getenv("USERPROFILE")

	} else {
		configDir = os.Getenv("HOME") + "/.config/hibercli"
		home = os.Getenv("HOME")

	}
	// créer le répertoire de configuration s'il n'existe pas
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, 0777)
	}

	// créer le fichier de configuration s'il n'existe pas
	configFilePath = configDir + "/config.yaml"
	vp := viper.New()

	//Définir le chemin de téléchargement par défaut
	_, err := os.Stat(home + "/Downloads") //Vérifier si le dossier Downloads existe
	if err != nil {
		dldPath = "./" //Si il n'existe pas on télécharge dans le dossier où se trouve l'utilisateur
	} else {
		dldPath = home + "/Downloads" //Sinon on télécharge dans le dossier Downloads
	}
	//Données par défaut du fichier de configuration
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
		"cli.notfound":  true,
		"cli.lastmsg":   time.Time{},
	}
	//Si le fichier de configuration n'existe pas on le crée et on y ajoute les données par défaut
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		os.Create(configFilePath)

		for key, value := range config {
			vp.Set(key, value)
		}

	}
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath(configDir)
	err = vp.ReadInConfig()
	if err != nil {
		red.Println(err)
		os.Exit(0)
	}
	var (
		currentDate    = time.Now()
		difference     = currentDate.Sub(vp.GetTime("cli.lastmsg"))
		currentVersion = "5.0.0"
	)
	//Obtenir la dernière version publiée
	resp, err := http.Get("https://api.github.com/repos/el2zay/hibercli/releases/latest")
	if err != nil {
		red.Println("Erreur : Impossible de récupérer la dernière version publiée\n", err)
		os.Exit(0)
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
		magenta.Println("\r", `
╭ Mise à jour ─────────────────────────────────────────────────╮
│                                                              │
│  Une nouvelle version est disponible`, red.Sprint(currentVersion), "→", green.Sprint(data["tag_name"]), magenta.Sprint(`          │
│    'hibercli set' pour activer les mises à jour automatiques │
│                                                              │
╰──────────────────────────────────────────────────────────────╯`))
		vp.Set("cli.lastmsg", currentDate)
	}
	//Ecrire dans la configuration
	err = vp.WriteConfig()
	if err != nil {
		red.Println("Erreur : Impossible d'écrire la configuration\n", err)
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
      hibercli [commande]

Commandes:
      download/d    Télécharger un fichier depuis HiberFile grâce à l'url du fichier
      help          Aide à propos d'une commande
      history       Affiche l'historique des fichiers téléversés
      issue         Ouvre une issue sur GitHub
      set/config    Paramétrer HiberCLI
      uninstall     Désinstaller HiberCLI
      upload/u      Téléverser un fichier sur HiberFile grâce au chemin du fichier
`)

}
