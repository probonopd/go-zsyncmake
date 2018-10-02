package main

import (
	"flag"
	"goZsyncmake/zsync"
	"goZsyncmake/zsyncOptions"

	"github.com/pkg/profile"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {

	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	}
	// 	if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	//opts := Options{0, "", "https://s3-us-west-2.amazonaws.com/zsync-benchmark/dummy.txt"}
	//zsyncMake("/home/agri/Documents/zsynctest/gozsync/go/dummy.txt", opts)

	// defer profile.Start(profile.MemProfile).Stop()
	defer profile.Start().Stop()

	opts := zsyncOptions.Options{0, "", ""}
	zsync.ZsyncMake("/home/agri/Downloads/ubuntu-18.04-live-server-amd64.iso", opts)

	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	runtime.GC() // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// 	f.Close()
	// }

}
