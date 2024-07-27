package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	searchLogsTable *tview.Table
	searchResult    []logGroup
	searchMenu      *tview.TextView
)

func createSearch() tview.Primitive {
	form := tview.NewForm()

	searchLogsTable = tview.NewTable()
	searchLogsTable.SetBorder(true)
	searchLogsTable.SetTitle("Search log groups")
	searchLogsTable.SetBackgroundColor(ColorBg)
	searchLogsTable.SetSelectable(true, false)
	searchLogsTable.Select(1, 0)
	searchLogsTable.SetCell(0, 0, tview.NewTableCell("Favorite").SetTextColor(ColorFg).SetSelectable(false))
	searchLogsTable.SetCell(0, 1, tview.NewTableCell("Name").SetTextColor(ColorFg).SetSelectable(false))

	form = tview.NewForm().
		AddInputField("Name", "", 0, nil, nil).
		AddButton("Search", func() {
			// get the name entry from the form
			name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
			searchResult = listCloudWatchGroups(name)
			if len(searchResult) > 0 {
				for i, group := range searchResult {
					// check allFavorites if log group is currently a favorite
					isFavorite := "No"
					for _, favorite := range allFavorites {
						if favorite.Name == group.Name {
							isFavorite = "Yes"
							break
						}
					}
					searchLogsTable.SetCell(i+1, 0, tview.NewTableCell(isFavorite).SetTextColor(ColorFg)).SetSelectable(true, false)
					searchLogsTable.SetCell(i+1, 1, tview.NewTableCell(group.Name).SetTextColor(ColorFg)).SetSelectable(true, false)
				}
				app.SetFocus(searchLogsTable)
			} else {
				searchLogsTable.SetCell(1, 0, tview.NewTableCell("- No log groups found").SetTextColor(ColorFg)).SetSelectable(true, false)
				app.SetFocus(form)
			}
		})

	form.SetBorder(true).SetTitle("Search log groups").SetTitleAlign(tview.AlignLeft)
	form.SetBackgroundColor(ColorBg)
	form.SetFieldTextColor(ColorInputFg)
	form.SetFieldBackgroundColor(ColorInputBg)
	form.SetButtonTextColor(ColorButtonFg)
	form.SetButtonBackgroundColor(ColorButtonBg)
	form.SetButtonActivatedStyle(tcell.StyleDefault.Foreground(ColorButtonFg).Background(ColorButtonActiveBg))

	// set focus function for Name input field
	nameInput := form.GetFormItemByLabel("Name").(*tview.InputField)
	nameInput.SetFocusFunc(func() {
		addConsoleRow("Search input focused, set hasActiveDialog to true")
		hasActiveDialog = true
	})
	nameInput.SetBlurFunc(func() {
		addConsoleRow("Search input blurred, set hasActiveDialog to false")
		hasActiveDialog = false
	})

	searchLogsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("search")
			pages.SwitchToPage(activePage)
			return nil
		}

		switch event.Rune() {
		case SearchAddFavoriteKey.KeyRune:
			addConsoleRow("Adding favorite")
			selectedRow, _ := searchLogsTable.GetSelection()
			logGroupName := searchLogsTable.GetCell(selectedRow, 1).Text
			for _, log := range searchResult {
				if log.Name == logGroupName {
					addFavorite(log)
					addConsoleRow("Added favorite")
					searchLogsTable.SetCell(selectedRow, 0, tview.NewTableCell("Yes").SetTextColor(ColorFg)).SetSelectable(true, false)
					break
				}
			}
			return nil
		case SearchOpenKey.KeyRune:
			selectedRow, _ := searchLogsTable.GetSelection()
			logGroup := searchLogsTable.GetCell(selectedRow, 1).Text
			addConsoleRow("Opening log group: " + logGroup)

			// get the arn from the selected row by looping favorites
			selectedArn := ""
			for _, log := range searchResult {
				if log.Name == logGroup {
					// remove last 2 chars from the ARN
					selectedArn = log.Arn[:len(log.Arn)-2]
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
			pages.RemovePage("search")
			go fetchLatestLogs(logGroup)
			return nil
		}

		return event
	})

	searchMenu = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)
	searchMenu.SetBackgroundColor(ColorSecondaryMenuBg)

	menuItems := [][]string{
		{SearchOpenKey.KeyLabel, SearchOpenKey.KeyDesc},
		{SearchAddFavoriteKey.KeyLabel, SearchAddFavoriteKey.KeyDesc},
	}
	var menuList []string

	for i := 0; i < len(menuItems); i++ {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", menuItems[i][0], menuItems[i][1]))
	}

	fmt.Fprintf(searchMenu, "%s", strings.Join(menuList, " "))

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 10, 1, true).
		AddItem(searchMenu, 1, 1, false).
		AddItem(searchLogsTable, 0, 1, false)

	return layout
}
