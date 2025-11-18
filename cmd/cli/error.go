package main

import "fmt"

func printError(err error) {
	if err == nil {
		return
	}
	fmt.Println("[ERROR]", err.Error())
}
