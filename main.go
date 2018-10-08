package main

import (
	"github.com/pkg/profile"
	"goZsyncmake/zsync"
)

func main() {

	//path := "/home/agri/Documents/ubuntu-18.04-live-server-amd64.iso"
	//file, err := os.Open(path)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//_, err = file.Stat()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//buff := make([]byte, 1024)
	//var tmpFile []byte
	//
	//for {
	//	_, err := file.Read(buff)
	//	if err != nil {
	//		if err == io.EOF {
	//			break
	//		}
	//		log.Fatal(err)
	//	}
	//
	//	if rand.Int() % 2 == 0 {
	//		tmpFile = append(tmpFile, buff...)
	//	}
	//
	//}
	//
	//file.Close()
	//
	//newFile, err := os.Create("/home/agri/Documents/zsynctest/ubuntu-18.04-live-server-amd64.iso")
	//newFile.Write(tmpFile)
	//newFile.Close()

	//opts := Options{0, "", "https://s3-us-west-2.amazonaws.com/zsync-benchmark/dummy.txt"}
	//zsyncMake("/home/agri/Documents/zsynctest/gozsync/go/dummy.txt", opts)

	//defer profile.Start(profile.MemProfile).Stop()
	defer profile.Start().Stop()

	opts := zsync.Options{0, "", "http://localhost:4572/zsynctesting/ubuntu-18.04-live-server-amd64.iso?AWSAccessKeyId=_not_needed_locally_&Signature=%2FwQK6McDqz4Feb6JLKVVEBFtpjA%3D&Expires=1538555571"}
	zsync.ZsyncMake("/home/agri/Documents/ubuntu-18.04-live-server-amd64.iso", opts)

}
