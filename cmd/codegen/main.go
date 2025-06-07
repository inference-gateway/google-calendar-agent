package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/inference-gateway/google-calendar-agent/internal/codegen"
)

var (
	output string
)

func init() {
	flag.StringVar(&output, "output", "", "Path to the output file")
}

func main() {
	flag.Parse()

	if output == "" {
		fmt.Println("Output must be specified")
		os.Exit(1)
	}

	fmt.Printf("Generating A2A types from schema to %s\n", output)
	err := codegen.GenerateA2ATypes(output, "a2a/a2a-schema.yaml")
	if err != nil {
		fmt.Printf("Error generating A2A types: %v\n", err)
		os.Exit(1)
	}
}
