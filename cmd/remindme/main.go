package main

import (
	"fmt"
	"remindme/internal/http"
	"time"
)

func main() {
	fmt.Println("it works!")
	http.StartApp()
	n := time.Now().UTC()
	fmt.Println(n)
}
