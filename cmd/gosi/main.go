package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/maildealru/gosi/pkg/config"
)

func main() {
	c, err := config.Parse(projDir())
	fmt.Printf("%+v, %s", c, err)
}

func projDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
