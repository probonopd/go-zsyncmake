package main

import (
	"github.com/pkg/profile"
	"goZsyncmake/zsync"
	"goZsyncmake/zsyncOptions"
)

func main() {

	//opts := Options{0, "", "https://s3-us-west-2.amazonaws.com/zsync-benchmark/dummy.txt"}
	//zsyncMake("/home/agri/Documents/zsynctest/gozsync/go/dummy.txt", opts)

	defer profile.Start(profile.MemProfile).Stop()
	//defer profile.Start().Stop()

	opts := zsyncOptions.Options{0, "", ""}
	zsync.ZsyncMake("/home/agri/Downloads/ubuntu-18.04-live-server-amd64.iso", opts)

}
