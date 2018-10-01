package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"go-zsyncmake/md4"
	"hash"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

func main() {
	opts := Options{0, "", ""}

	zsyncMake("/home/agri/Documents/zsynctest/gozsync/go/dummy.txt", opts)
}

func zsyncMake(path string, options Options) {
	checksum, headers, zsyncFilePath := writeToFile(path, options)
	zsyncFile, err := os.Create(zsyncFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer zsyncFile.Close()

	bfio := bufio.NewWriter(zsyncFile)
	_, err = bfio.WriteString(headers)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bfio.Write(checksum)
	if err != nil {
		log.Fatal(err)
	}

	bfio.Flush()

	//_, err = zsyncFile.WriteString(headers)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//_, err = zsyncFile.Write(checksum)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = zsyncFile.Sync()
	//if err != nil {
	//	log.Fatal(err)
	//}
}

var ZSYNC_VERSION = "0.6.2"
func writeToFile(path string, options Options) ([]byte, string, string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	outputFileName := file.Name() + ".zsync"
	//println("outputFileName: " + outputFileName)

	fileInfo, err := file.Stat();
	if(err != nil) {
		log.Fatal(err)
	}

	opts := calculateMissingValues(options, file)

	blockSize := opts.blockSize
	fileLength := fileInfo.Size()
	sequenceMatches := 0
	if(fileLength > int64(options.blockSize)) {
		sequenceMatches = 2
	} else {
		sequenceMatches = 1
	}
	weakChecksumLength := weakChecksumLength(fileLength, blockSize, sequenceMatches)
	strongChecksumLength := strongChecksumLength(fileLength, blockSize, sequenceMatches)

	fileDigest := sha1.New()
	blockDigest := md4.New()	// should be imported from golib, I do quick hack by localize it

	checksum, fileChecksum := computeChecksum(file, blockSize, fileLength, weakChecksumLength, strongChecksumLength, fileDigest, blockDigest)
	strFileChecksum := hex.EncodeToString(fileChecksum)

	strHeader := "zsync: " + ZSYNC_VERSION + "\n" +
		"Filename: " + file.Name() + "\n" +
		"MTime: " + strconv.Itoa(int(fileInfo.ModTime().Unix())) + "\n" +
		"Blocksize: " + strconv.Itoa(blockSize) + "\n" +
		"Length: " + strconv.Itoa(int(fileLength)) + "\n" +
		"Hash-Lengths: " + strconv.Itoa(sequenceMatches) + "," + strconv.Itoa(weakChecksumLength) + "," + strconv.Itoa(strongChecksumLength)+ "\n" +
		"URL: " + opts.url + "\n" +
		"SHA-1: " + strFileChecksum + "\n\n"

	return checksum, strHeader, outputFileName


}

func computeChecksum(f *os.File, blocksize int, fileLength int64, weakLen int, strongLen int, fileDigest hash.Hash, blockDigest hash.Hash) ([]byte, []byte) {
	a := fileLength / int64(blocksize)
	b := int64(0)
	if(fileLength % int64(blocksize) > 0) {
		b = int64(1)
	}

	capacity := (a + b) * int64(weakLen + strongLen) + int64(fileDigest.Size());
	println(capacity)

	checksumBytes := make([]byte, 0)
	block := make([]byte, blocksize)
	wholeBlockFile := make([]byte, 0)

	for {
		read, err := f.Read(block)
		if(err != nil) {
			if(err == io.EOF) {
				break
			}
			log.Fatal(err)
		}

		//encode := base64.StdEncoding.EncodeToString(block)
		//println(encode)

		if(read < blocksize) {

			for i := 0; i < read; i++ {
				wholeBlockFile = append(wholeBlockFile, block[i])
			}

			blockSlice := block[read:blocksize]
			for i := range blockSlice {
				blockSlice[i] = byte(0)
			}

		} else {
			wholeBlockFile = append(wholeBlockFile, block...)
		}

		rsum := computeRsum(block)

		_, unsignedWeakByte := intToByteArr(int32(rsum))
		strbase64(unsignedWeakByte)

		tempUnsignedWeakByte := unsignedWeakByte[len(unsignedWeakByte) - weakLen:]
		checksumBytes = append(checksumBytes, tempUnsignedWeakByte...)

		blockDigest.Reset()
		blockDigest.Write(block)
		strongBytes := blockDigest.Sum(nil)
		strbase64(strongBytes)

		//signedInts, unsignedStrong := calculateSignedByte(strongBytes)
		_, unsignedStrong := calculateSignedByte(strongBytes)

		//print(signedInts, unsignedStrong)

		tempUnsignedStrongByte := unsignedStrong[:strongLen]
		checksumBytes = append(checksumBytes, tempUnsignedStrongByte...)

		//println("")

	}

	//if _, err := io.Copy(fileDigest, f); err != nil {
	//	log.Fatal(err)
	//}

	fileDigest.Reset()
	fileDigest.Write(wholeBlockFile)
	fileChecksum := fileDigest.Sum(nil)

	//signedFileChecksumInts, unsignedFileChecksumBytes := calculateSignedByte(fileChecksum)
	_, unsignedFileChecksumBytes := calculateSignedByte(fileChecksum)

	print("filechecksum sha1: ")
	strbase64(unsignedFileChecksumBytes)

	//println(signedFileChecksumInts, unsignedFileChecksumBytes)

	// TODO change unsignedFileChecksumBytes to fileChecksum and remove calculateSignedByte, this case unnecesary
	checksumBytes = append(checksumBytes, unsignedFileChecksumBytes...)


	return checksumBytes, fileChecksum

}

func strbase64(bl []byte) {
	encode := base64.StdEncoding.EncodeToString(bl)
	println(encode)
}

func intToByteArr(v int32) ([]int8, []byte) {

	unsigned := []byte {
		byte((v >> 24) & 0xff),
		byte((v >> 16) & 0xff),
		byte((v >> 8) & 0xff),
		byte((v >> 0) & 0xff),
	}

	return calculateSignedByte(unsigned)
}

func calculateSignedByte(unsigned []byte) ([]int8, []byte) {
	signed := make([]int8, len(unsigned))
	unsignedByte := make([]byte, len(unsigned))
	for i, v := range unsigned {
		signed[i] = int8(v)
		unsignedByte[i] = byte(signed[i])
	}

	return signed, unsignedByte
}

func strongChecksumLength(fileLength int64, blocksize int, sequenceMatches int) int {
	// estimated number of bytes to allocate for strong checksum
	d := (math.Log(float64(fileLength)) + math.Log(float64(1 + fileLength / int64(blocksize)))) / math.Log(2) + 20

	// reduced number of bits by sequence matches
	lFirst := float64(math.Ceil(d / float64(sequenceMatches) / 8))

	// second checksum - not reduced by sequence matches
	lSecond := float64((math.Log(float64(1 + fileLength / int64(blocksize))) / math.Log(2) + 20 + 7.9) / 8);

	// return max of two: return no more than 16 bytes (MD4 max)
	return int(math.Min(float64(16), math.Max(lFirst, lSecond)))
}

func weakChecksumLength(fileLength int64, blocksize int, sequenceMatches int) int {
	// estimated number of bytes to allocate for the rolling checksum per formula in
	// Weak Checksum section of http://zsync.moria.org.uk/paper/ch02s03.html
	d := (math.Log(float64(fileLength)) + math.Log(float64(blocksize))) / math.Log(2) - 8.6

	// reduced number of bits by sequence matches per http://zsync.moria.org.uk/paper/ch02s04.html
	rdc := d / float64(sequenceMatches) / 8
	lrdc := int(math.Ceil(rdc))

	// enforce max and min values
	if(lrdc > 4) {
		return 4
	} else {
		if(lrdc < 2) {
			return 2
		} else {
			return lrdc
		}
	}
}

func computeRsum(block []byte) int {
	var a int16
	var b int16
	l := len(block)
	for i := 0; i < len(block); i++ {
		val := int(unsign(block[i]))
		a += int16(val)
		b += int16(l * val)
		l--
	}
	x := int(a) << 16
	y := int(b) & 0xffff
	return int(x) | int(y)
}

func unsign(b byte) uint8 {
	if b < 0  {
		return b & 0xFF
	} else {
		return b
	}
}


type Options struct {
	blockSize int
	filename string
	url string
}

func calculateMissingValues(opts Options, f *os.File) Options {
	if(opts.blockSize == 0) {
		opts.blockSize = calculateDefaultBlockSizeForInputFile(f)
	}
	if(opts.filename == "") {
		opts.filename = f.Name()
	}
	if(opts.url == "") {
		opts.url = f.Name()
	}
	return opts
}

var BLOCK_SIZE_SMALL = 2048
var BLOCK_SIZE_LARGE = 4096
func calculateDefaultBlockSizeForInputFile(f *os.File) int {
	fileInfo, err := f.Stat()
	if(err != nil) {
		log.Fatal(err)
	}
	if(fileInfo.Size() < 100 * 1 << 20) {
		return BLOCK_SIZE_SMALL
	} else {
		return BLOCK_SIZE_LARGE
	}
}