package main

import (
	"ChaikaReports/internal/config"
	"fmt"
)

func main() {
	cfg := config.LoadConfig("config.yml")
	fmt.Println(cfg)
}
