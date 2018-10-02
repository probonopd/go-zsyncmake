package main

import (
	"goZsyncmake/zsync"
	"goZsyncmake/zsyncOptions"
	"testing"
)

func TestZsyncMake(t *testing.T) {
	opts := zsyncOptions.Options{0, "", ""}
	zsync.ZsyncMake("/home/agri/Downloads/ubuntu-18.04-live-server-amd64.iso", opts)
}
