package main

import (
	"fmt"

	"github.com/rivo/tview"
)

var consoleTable *tview.Table

func createConsole() tview.Primitive {
	consoleTable = tview.NewTable()
	consoleTable.SetBorder(true)
	consoleTable.SetTitle("Console")
	consoleTable.SetBackgroundColor(ColorConsoleBg)

	return consoleTable
}

func addConsoleRow(text string) {
	if consoleTable == nil {
		fmt.Println("consoleTable is nil")
		return
	}
	rowCount := consoleTable.GetRowCount()
	consoleTable.SetCell(rowCount, 0, tview.NewTableCell(text).SetTextColor(ColorFg).SetSelectable(false))
}
