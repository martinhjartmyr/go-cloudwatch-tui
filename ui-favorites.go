package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

type logGroup struct {
	Name    string `json:"name"`
	Arn     string `json:"arn"`
	Profile string `json:"profile"`
}

var (
	favoritesTable  *tview.Table
	favoritesMenu   *tview.TextView
	allFavorites    []logGroup
	activeFavorites []logGroup
)

func createFavorites(favorites []logGroup) tview.Primitive {
	allFavorites = favorites
	activeFavorites = favorites
	favoritesTable = tview.NewTable()
	favoritesTable.SetBackgroundColor(ColorBg)
	favoritesTable.SetSelectable(true, false)
	favoritesTable.Select(1, 0)
	favoritesTable.SetCell(0, 0, tview.NewTableCell("Profile").SetTextColor(ColorFg).SetSelectable(false))
	favoritesTable.SetCell(0, 1, tview.NewTableCell("Name").SetTextColor(ColorFg).SetExpansion(1).SetSelectable(false))

	favoritesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case FavoriteAddKey.KeyRune:
			if !pages.HasPage("search") {
				pages.AddAndSwitchToPage("search", createSearch(), true)
			} else {
				pages.SwitchToPage("search")
			}
			return nil
		case FavoriteOpenKey.KeyRune:
			selectedRow, _ := favoritesTable.GetSelection()
			logGroup := favoritesTable.GetCell(selectedRow, 1).Text
			addConsoleRow("Opening log group: " + logGroup)

			// get the arn from the selected row by looping favorites
			selectedArn := ""
			for _, favorite := range favorites {
				if favorite.Name == logGroup {
					// remove last 2 chars from the ARN
					selectedArn = favorite.Arn[:len(favorite.Arn)-2]
				}
			}

			if selectedArn == "" {
				addConsoleRow("No ARN found for log group: " + logGroup)
				return event
			} else {
				addConsoleRow("ARN found for log group: " + selectedArn)
				activeLogGroup = selectedArn
				drawUILogsMenu()
			}
			pages.SwitchToPage("logs")
			go fetchLatestLogs(logGroup)
			return nil
		case FavoriteDeleteKey.KeyRune:
			selectedRow, _ := favoritesTable.GetSelection()
			logGroupName := favoritesTable.GetCell(selectedRow, 1).Text

			var selectedLogGroup logGroup
			for _, favorite := range activeFavorites {
				if favorite.Name == logGroupName {
					selectedLogGroup = favorite
				}
			}
			deleteFavorite(selectedLogGroup.Profile, selectedLogGroup.Arn)
			saveFavorites()
			loadFavorites(selectedRow)
			return nil
		}

		return event
	})

	favoritesMenu = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)
	favoritesMenu.SetBackgroundColor(ColorSecondaryMenuBg)

	menuItems := [][]string{
		{FavoriteAddKey.KeyLabel, FavoriteAddKey.KeyDesc},
		{FavoriteOpenKey.KeyLabel, FavoriteOpenKey.KeyDesc},
		{FavoriteDeleteKey.KeyLabel, FavoriteDeleteKey.KeyDesc},
	}
	var menuList []string

	for i := 0; i < len(menuItems); i++ {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", menuItems[i][0], menuItems[i][1]))
	}

	fmt.Fprintf(favoritesMenu, "%s", strings.Join(menuList, " "))

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(favoritesMenu, 1, 1, false).
		AddItem(favoritesTable, 0, 1, true)

	return layout
}

func loadFavorites(selectRow int) {
	// TODO: do filter stuff
	activeFavorites = []logGroup{}
	for _, favorite := range allFavorites {
		if favorite.Profile == activeAwsProfile {
			activeFavorites = append(activeFavorites, favorite)
		}
	}

	addConsoleRow("Loading favorites")
	favoritesTable.Clear()
	favoritesTable.SetCell(0, 0, tview.NewTableCell("Profile").SetTextColor(ColorFg).SetSelectable(false))
	favoritesTable.SetCell(0, 1, tview.NewTableCell("Name").SetTextColor(ColorFg).SetExpansion(1).SetSelectable(false))
	for i, favorite := range activeFavorites {
		favoritesTable.SetCell(i+1, 0, tview.NewTableCell(favorite.Profile).SetTextColor(ColorFg)).SetSelectable(true, false)
		favoritesTable.SetCell(i+1, 1, tview.NewTableCell(favorite.Name).SetTextColor(ColorFg)).SetSelectable(true, false)
	}
	if selectRow > 0 {
		favoritesTable.Select(selectRow, 0)
	} else {
		favoritesTable.Select(1, 0)
	}
}

func saveFavorites() {
	viper.Set("favorites", allFavorites)
	viper.WriteConfig()
	addConsoleRow("Saved favorites")
}

func deleteFavorite(profile string, arn string) {
	// remove favorite from allFavorites
	for i, favorite := range allFavorites {
		if favorite.Profile == profile && favorite.Arn == arn {
			allFavorites = append(allFavorites[:i], allFavorites[i+1:]...)
		}
	}
}

func addFavorite(newLogGroup logGroup) {
	allFavorites = append(allFavorites, newLogGroup)
	viper.Set("favorites", allFavorites)
	viper.WriteConfig()
	addConsoleRow("Saved favorites")
}
