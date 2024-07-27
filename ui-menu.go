package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func createMenu(menuItems [][]string) *tview.TextView {
	menu := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)
	menu.SetBackgroundColor(ColorMenuBg)

	var menuList []string
	for i := 0; i < len(menuItems); i++ {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", menuItems[i][0], menuItems[i][1]))
	}

	fmt.Fprintf(menu, "%s", strings.Join(menuList, " "))

	return menu
}
