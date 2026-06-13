package main

import (
	"app/budget-planner/config"
	"shared/loggers"
)

func main() {
	logger := loggers.NewLogger()
	//
	conf := config.NewConfig(logger)

}
