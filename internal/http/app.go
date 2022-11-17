package http

import (
	"fmt"
	"remindme/internal/domain/service"
)

func StartApp() {
	fmt.Println("App is running...")
	fmt.Println(service.PrintUser())
}
