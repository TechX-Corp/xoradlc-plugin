package main

import (
	"fmt"
	"os"

	"github.com/TechX-Corp/xoradlc-plugin/internal/app"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
)

func main() {
	err := app.Run(os.Args[1:])
	if err != nil {
		if ab, ok := err.(*cli.Abort); ok {
			if ab.Msg != "" {
				if ab.Code == 0 {
					fmt.Println(ab.Msg)
				} else {
					fmt.Fprintf(os.Stderr, "xdlc: %s\n", ab.Msg)
				}
			}
			os.Exit(ab.Code)
		}
		fmt.Fprintf(os.Stderr, "xdlc: %v\n", err)
		os.Exit(1)
	}
}