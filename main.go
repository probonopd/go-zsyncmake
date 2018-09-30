package main

import (
	"crypto/sha1"
	"encoding/base64"
	"hash"
	"log"
	"math"
	"os"
	"zsyncMake/md4"
)

func main() {
	opts := Options{0, "", ""}

	zsyncMake("C:\\Users\\root\\Documents\\Accelbyte\\dummy.txt", opts)
}

func zsyncMake(path string, options Options) {
	writeToFile(path, options)
}


func writeToFile(path string, options Options) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	outputFileName := file.Name() + ".zsync"
	println("outputFileName: " + outputFileName)

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

	println(weakChecksumLength)
	println(strongChecksumLength)

	fileDigest := sha1.New()
	println(fileDigest.Size())
	blockDigest := md4.New()	// should be imported from golib, I do quick hack by localize it
	println(blockDigest.Size())

	computeChecksum(file, blockSize, fileLength, weakChecksumLength, strongChecksumLength, fileDigest, blockDigest)


}

func computeChecksum(f *os.File, blocksize int, fileLength int64, weakLen int, strongLen int, fileDigest hash.Hash, blockDigest hash.Hash) {
	//a := fileLength / int64(blocksize)
	//b := int64(0)
	//if(fileLength % int64(blocksize) > 0) {
	//	b = int64(1)
	//}

	//capacity := (a + b) * int64(weakLen + strongLen) + int64(fileDigest.Size());

	checksumBytes := make([]byte, 0)
	block := make([]byte, blocksize)

	for {
		read, err := f.Read(block)
		if(err != nil) {
			log.Fatal(err)
		}

		//encode := base64.StdEncoding.EncodeToString(block)
		//println(encode)

		if(read < blocksize) {
			blockSlice := block[read:blocksize]
			for i := range blockSlice {
				blockSlice[i] = 0
			}
			break
		}

		rsum := computeRsum(block)

		signedWeakInts, unsignedWeakByte := intToByteArr(int32(rsum))

		println(signedWeakInts)
		println("")
		strbase64(unsignedWeakByte)
		//bs := new(bytes.Buffer)
		//b := make([]byte, 4)
		//binary.BigEndian.PutUint32(b, uint32(rsum))
		//
		//err = binary.Write(bs, binary.BigEndian, int32(rsum))
		//if err != nil {
		//	log.Fatal(err)
		//}
		//bytearr := bs.Bytes()
		//println(bytearr)

		//rsum32 := uint32(rsum)
		//b := make([]byte, 4)
		//binary.BigEndian.PutUint32(b, rsum32)
		//binary.LittleEndian.PutUint32(b, uint32(rsum))
		//varint := binary.PutVarint(b, int64(rsum))
		//println(varint)

		checksumBytes = append(checksumBytes, unsignedWeakByte...)

		//blockDigest.Write(signedByte)
		strongBytes := blockDigest.Sum(unsignedWeakByte)

		println("")
		strbase64(strongBytes)

		signedInts, signedStrong := calculateSignedByte(strongBytes)

		print(signedInts,signedStrong)

		checksumBytes = append(checksumBytes, strongBytes...)

		println("")

	}

	println(len(checksumBytes))

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