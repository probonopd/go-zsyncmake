package zsync

import (
	"bufio"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"goZsyncmake/md4"
	"goZsyncmake/zsyncOptions"
	"hash"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

var ZSYNC_VERSION = "0.6.2"
var BLOCK_SIZE_SMALL = 2048
var BLOCK_SIZE_LARGE = 4096

// ZsyncMake zsync make
func ZsyncMake(path string, options zsyncOptions.Options) {
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
}

func writeToFile(path string, options zsyncOptions.Options) ([]byte, string, string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	outputFileName := file.Name() + ".zsync"

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	opts := calculateMissingValues(options, file)

	blockSize := opts.BlockSize
	fileLength := fileInfo.Size()
	sequenceMatches := 0
	if fileLength > int64(options.BlockSize) {
		sequenceMatches = 2
	} else {
		sequenceMatches = 1
	}
	weakChecksumLength := weakChecksumLength(fileLength, blockSize, sequenceMatches)
	strongChecksumLength := strongChecksumLength(fileLength, blockSize, sequenceMatches)

	fileDigest := sha1.New()
	blockDigest := md4.New() // should be imported from golib, I do quick hack by localize it

	checksum, fileChecksum := computeChecksum(file, blockSize, fileLength, weakChecksumLength, strongChecksumLength, fileDigest, blockDigest)
	strFileChecksum := hex.EncodeToString(fileChecksum)

	strHeader := "zsync: " + ZSYNC_VERSION + "\n" +
		"Filename: " + fileInfo.Name() + "\n" +
		"MTime: " + fileInfo.ModTime().Format(time.RFC1123Z) + "\n" +
		"Blocksize: " + strconv.Itoa(blockSize) + "\n" +
		"Length: " + strconv.Itoa(int(fileLength)) + "\n" +
		"Hash-Lengths: " + strconv.Itoa(sequenceMatches) + "," + strconv.Itoa(weakChecksumLength) + "," + strconv.Itoa(strongChecksumLength) + "\n" +
		"URL: " + opts.Url + "\n" +
		"SHA-1: " + strFileChecksum + "\n\n"

	return checksum, strHeader, outputFileName

}

func sha1HashFile(path string, fileChecksumChannel chan []byte) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Fatal(err)
	}

	fileChecksumChannel <- hasher.Sum(nil)
}

func computeChecksum(f *os.File, blocksize int, fileLength int64, weakLen int, strongLen int, fileDigest hash.Hash, blockDigest hash.Hash) ([]byte, []byte) {

	checksumBytes := make([]byte, 0)
	block := make([]byte, blocksize)
	//wholeBlockFile := make([]byte, 0)

	fileChecksumChannel := make(chan []byte)
	go sha1HashFile(f.Name(), fileChecksumChannel)

	for {
		read, err := f.Read(block)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		if read < blocksize {

			//for i := 0; i < read; i++ {
			//	wholeBlockFile = append(wholeBlockFile, block[i])
			//}

			blockSlice := block[read:blocksize]
			for i := range blockSlice {
				blockSlice[i] = byte(0)
			}

		} else {
			//wholeBlockFile = append(wholeBlockFile, block...)
		}

		rsum := computeRsum(block)

		unsignedWeakByte := make([]byte, 4)
		binary.BigEndian.PutUint32(unsignedWeakByte, uint32(rsum))

		tempUnsignedWeakByte := unsignedWeakByte[len(unsignedWeakByte)-weakLen:]
		checksumBytes = append(checksumBytes, tempUnsignedWeakByte...)

		blockDigest.Reset()
		blockDigest.Write(block)
		strongBytes := blockDigest.Sum(nil)

		tempUnsignedStrongByte := strongBytes[:strongLen]
		checksumBytes = append(checksumBytes, tempUnsignedStrongByte...)

	}

	//fileDigest.Reset()
	//fileDigest.Write(wholeBlockFile)
	//fileChecksum := fileDigest.Sum(nil)

	fileChecksum := <- fileChecksumChannel


	// TODO change unsignedFileChecksumBytes to fileChecksum and remove calculateSignedByte, this case unnecesary
	checksumBytes = append(checksumBytes, fileChecksum...)

	return checksumBytes, fileChecksum

}

func strongChecksumLength(fileLength int64, blocksize int, sequenceMatches int) int {
	// estimated number of bytes to allocate for strong checksum
	d := (math.Log(float64(fileLength))+math.Log(float64(1+fileLength/int64(blocksize))))/math.Log(2) + 20

	// reduced number of bits by sequence matches
	lFirst := float64(math.Ceil(d / float64(sequenceMatches) / 8))

	// second checksum - not reduced by sequence matches
	lSecond := float64((math.Log(float64(1+fileLength/int64(blocksize)))/math.Log(2) + 20 + 7.9) / 8)

	// return max of two: return no more than 16 bytes (MD4 max)
	return int(math.Min(float64(16), math.Max(lFirst, lSecond)))
}

func weakChecksumLength(fileLength int64, blocksize int, sequenceMatches int) int {
	// estimated number of bytes to allocate for the rolling checksum per formula in
	// Weak Checksum section of http://zsync.moria.org.uk/paper/ch02s03.html
	d := (math.Log(float64(fileLength))+math.Log(float64(blocksize)))/math.Log(2) - 8.6

	// reduced number of bits by sequence matches per http://zsync.moria.org.uk/paper/ch02s04.html
	rdc := d / float64(sequenceMatches) / 8
	lrdc := int(math.Ceil(rdc))

	// enforce max and min values
	if lrdc > 4 {
		return 4
	} else {
		if lrdc < 2 {
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
	if b < 0 {
		return b & 0xFF
	} else {
		return b
	}
}

func calculateMissingValues(opts zsyncOptions.Options, f *os.File) zsyncOptions.Options {
	if opts.BlockSize == 0 {
		opts.BlockSize = calculateDefaultBlockSizeForInputFile(f)
	}
	if opts.Filename == "" {
		opts.Filename = f.Name()
	}
	if opts.Url == "" {
		opts.Url = f.Name()
	}
	return opts
}

func calculateDefaultBlockSizeForInputFile(f *os.File) int {
	fileInfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.Size() < 100*1<<20 {
		return BLOCK_SIZE_SMALL
	} else {
		return BLOCK_SIZE_LARGE
	}
}
