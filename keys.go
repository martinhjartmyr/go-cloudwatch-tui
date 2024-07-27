package main

import "github.com/gdamore/tcell/v2"

type keyMapping struct {
	Key      tcell.Key
	KeyRune  rune
	KeyLabel string
	KeyDesc  string
}

var (
	QuitKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('q'),
		KeyLabel: "Q",
		KeyDesc:  "Quit",
	}
	HelpScreenKey = keyMapping{
		Key:      tcell.KeyF1,
		KeyLabel: "F1",
		KeyDesc:  "Help",
	}
	FavoritesScreenKey = keyMapping{
		Key:      tcell.KeyF2,
		KeyLabel: "F2",
		KeyDesc:  "Favorites",
	}
	LogsScreenKey = keyMapping{
		Key:      tcell.KeyF3,
		KeyLabel: "F3",
		KeyDesc:  "Logs",
	}
	FavoriteOpenKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('o'),
		KeyLabel: "O",
		KeyDesc:  "Open",
	}
	FavoriteAddKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('a'),
		KeyLabel: "A",
		KeyDesc:  "Add",
	}
	FavoriteSaveKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('s'),
		KeyLabel: "S",
		KeyDesc:  "Save",
	}
	FavoriteDeleteKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('d'),
		KeyLabel: "D",
		KeyDesc:  "Delete",
	}
	LogsTailKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('t'),
		KeyLabel: "T",
		KeyDesc:  "Tail",
	}
	LogsChangeKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('l'),
		KeyLabel: "L",
		KeyDesc:  "Tail",
	}
	LogsRefreshKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('r'),
		KeyLabel: "R",
		KeyDesc:  "Refresh",
	}
	SearchAddFavoriteKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('a'),
		KeyLabel: "A",
		KeyDesc:  "Add favorite",
	}
	SearchOpenKey = keyMapping{
		Key:      tcell.Key(256),
		KeyRune:  rune('o'),
		KeyLabel: "O",
		KeyDesc:  "Open log group",
	}
	ProfileScreenKey = keyMapping{
		Key:      tcell.KeyF1,
		KeyLabel: "F1",
		KeyDesc:  "Profile",
	}
)
