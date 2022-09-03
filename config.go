package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type confItem struct {
	Res         *string
	Default     string
	PanicIfNone bool
}


var (
	DB_URL = ""
)

func InitConfig() {
	godotenv.Load(".env")

	var confMap = map[string]confItem{
		"DB_URL": {
			Res: &DB_URL,
			PanicIfNone: true,
		},
	}

	for name, opt := range confMap {
		item := os.Getenv(name)
		
		if len(item) == 0 {
			if opt.PanicIfNone {
				panic(fmt.Sprintf("'%v' is a needed variable, but is not present! Please read the README.md file for more info.", name))
			}
			item = opt.Default
		}

		*opt.Res = item
	}
}
