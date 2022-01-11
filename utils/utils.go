package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
)

func WaitForCtrlC() {
	quit := make(chan bool)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	go func() {
		for range sigint {
			quit <- true
		}
	}()

	<-quit
}

func GetRandomIntRange(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func FailOnError(err error, message string) {
	if err != nil {
		errorMessage := fmt.Sprintf("%s: %s", message, err)

		panic(errorMessage)
	}
}
