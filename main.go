package main

import (
	"fmt"
	"os"
)

func main() {
    fmt.Println("File to check:", os.Getenv("INPUT_FILEPATH"))
}
