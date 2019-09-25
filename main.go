package main

import (
	"fmt"

	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/models"
)

func main() {
	fmt.Println("Hello")
	models.ConnectDB()

	handler.Restapi()
}
