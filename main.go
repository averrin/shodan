package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func main() {
	log.Println("======")
	nobot = flag.Bool("nobot", false, "no telegram bot")
	flag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	shodan := NewShodan()
	shodan.Serve()
}
