package main

import (
	"github.com/hupe1980/zipbomb/cmd"
)

var (
	version = "dev"
)

func main() {
	cmd.Execute(version)
}
