package main

import (
	"fmt"
	"wv2/types"
	"wv2/widgets"
)

// In prod, this will be from some source
var botList = []types.WidgetUser{
	{
		ID:       "564164277251080208",
		Username: "selectthegang",
		Avatar:   "https://cdn.discordapp.com/avatars/564164277251080208/7cb4cdee538d9621aa80e14bfed106d8.webp?size=1024",
		OutFile:  "tmp/select.webp",
	},
}

func main() {
	for _, bot := range botList {
		fmt.Println(bot.Username)
		err := bot.ParseData()
		if err != nil {
			fmt.Println(err)
		}
		widgets.DrawWidget(bot)
	}
}
