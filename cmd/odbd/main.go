// Package odbd implements the server component of a odb server
package main

import (
	"flag"
)

var (
	registryFolder = flag.String("databaseFile", "registry.bolt", "Folder to keep the files")
)

func main() {
	NewServer(*registryFolder)
}