package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

func createProfileModal(profiles []string) tview.Primitive {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetTitle("Select AWS Profile")
	table.SetSelectable(true, false)
	table.Select(0, 0)

	for i, profile := range profiles {
		table.SetCell(i, 0, tview.NewTableCell(profile).SetTextColor(ColorFg)).SetSelectable(true, false)
	}

	table.SetSelectedFunc(func(row, column int) {
		value := table.GetCell(row, 0).Text
		viper.Set("aws.profile", value)
		viper.WriteConfig()
		activeAwsProfile = value

		awsConfigLoaded, errConfig := loadAWSConfig(activeAwsProfile)
		awsConfig = awsConfigLoaded

		if errConfig != nil {
			fmt.Println(errConfig)
			panic(errConfig)
		}

		loadFavorites(0)
		pages.RemovePage("modal")
		pages.SwitchToPage(activePage)
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("modal")
			pages.SwitchToPage(activePage)
		}

		return event
	})

	return table
}
