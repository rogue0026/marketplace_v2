package main

import (
	"context"
	"fmt"
	"order_service/internal/application"
)

func main() {
	ctx := context.Background()
	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = app.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
