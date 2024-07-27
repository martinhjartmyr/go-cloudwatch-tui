package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

var (
	app              *tview.Application
	pages            *tview.Pages
	menu             *tview.TextView
	status           *tview.TextView
	activePage       string
	activeModal      *tview.Primitive
	activeAwsProfile string
	hasActiveDialog  bool
	awsConfig        aws.Config
	showConsole      bool
)

func main() {
	hasActiveDialog = false
	showConsole = false
	activePage = "favorites"

	// check for command line arguments
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Println("Usage: gct [options]")
			fmt.Println("  -h, --help       Show this help message")
			fmt.Println("  -c, --console    Show debug console")
			os.Exit(0)
		case "-c", "--console":
			showConsole = true
		default:
			fmt.Println("Unknown argument: ", arg)
			os.Exit(1)
		}
	}

	loadConfig()

	awsProfiles, errProfiles := listAwsProfiles()
	awsProfiles = append([]string{"default"}, awsProfiles...)
	activeAwsProfile = viper.GetString("aws.profile")
	awsConfigLoaded, errConfig := loadAWSConfig(activeAwsProfile)
	awsConfig = awsConfigLoaded
	if errConfig != nil || errProfiles != nil {
		fmt.Println(awsProfiles)
		panic(errConfig)
	}

	app = tview.NewApplication()

	menu = createMenu([][]string{
		{ProfileScreenKey.KeyLabel, fmt.Sprintf("%s:%s", ProfileScreenKey.KeyDesc, activeAwsProfile)},
		{FavoritesScreenKey.KeyLabel, FavoritesScreenKey.KeyDesc},
		{LogsScreenKey.KeyLabel, LogsScreenKey.KeyDesc},
		{QuitKey.KeyLabel, QuitKey.KeyDesc},
	})

	// load favorites from config
	favGroups := []logGroup{}
	if viper.IsSet("favorites") {
		if err := viper.UnmarshalKey("favorites", &favGroups); err != nil {
			panic(err)
		}
	}

	console := createConsole()

	pages = tview.NewPages().
		AddPage("favorites", createFavorites(favGroups), true, false).
		AddPage("logs", createLogs(), true, false)

	pages.SetBackgroundColor(ColorBg)

	loadFavorites(1)
	pages.SwitchToPage(activePage)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !hasActiveDialog {
			switch event.Rune() {
			case QuitKey.KeyRune:
				app.Stop()
				return nil
			}

			switch event.Key() {
			case ProfileScreenKey.Key:
				profileModal := createProfileModal(awsProfiles)
				pages.AddAndSwitchToPage("modal", createModal(profileModal, 60, 20), true)
				return nil
			case FavoritesScreenKey.Key:
				pages.SwitchToPage("favorites")
				activePage = "favorites"
				loadFavorites(1)
				return nil
			case LogsScreenKey.Key:
				pages.SwitchToPage("logs")
				activePage = "logs"
				return nil
			}

		}

		return event
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(menu, 1, 1, false).
		AddItem(pages, 0, 3, true)

	if showConsole {
		flex.AddItem(console, 0, 1, false)
	}

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func createModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func loadConfig() {
	userHome, errHomeDir := os.UserHomeDir()
	if errHomeDir != nil {
		panic(errHomeDir)
	}
	configDir := fmt.Sprintf("%s/.config/go-cloudwatch-tui", userHome)
	configFile := fmt.Sprintf("%s/config.json", configDir)
	viper.SetConfigType("json")
	viper.SetConfigFile(configFile)
	viper.SetDefault("aws.profile", "default")
	errConfig := viper.ReadInConfig()

	if errConfig != nil {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println("Failed to create config directory", err)
		}
		if err := viper.WriteConfigAs(configFile); err != nil {
			fmt.Println("Failed to save config", err)
		}
	}
}
