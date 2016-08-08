package main

import (
	"flag"
	"fmt"
	"log"

	// at "./modules/attendance/"

	"github.com/spf13/viper"
)

func main() {
	log.Println("======")
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
