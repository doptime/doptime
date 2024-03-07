package goflow

import (
	"fmt"
	"time"
)

func Start() {
	fmt.Println("GoFlow Has started successfully!")
	//sleep always
	for {
		time.Sleep(time.Second * 10)
	}
}
