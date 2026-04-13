package main

import (
	"fmt"
	"gateway/internal/application"
)

func main() {
	app, err := application.New("./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("starting listening")
	err = app.Server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
