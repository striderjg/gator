package main

import (
	"fmt"

	"github.com/striderjg/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := cfg.SetUser("John"); err != nil {
		fmt.Println(err.Error())
		return
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(cfg.Current_user)
	fmt.Println(cfg.Db_url)
}
