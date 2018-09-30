package main
//
//var BLOCK_SIZE_SMALL = 2048;
//var BLOCK_SIZE_LARGE = 4096;
//
//type Options struct {
//
//	blockSize int
//	filename string
//	url string
//
//	/**
//	 * Resolves option values which are required for the zsyncmake operation but which were not supplied.
//	 */
//
//}
//
//func calculateMissingValues(inputFile Path, options Options) Options {
//	// blocksize: default chosen based on file size (adopted from standard zsync implementation)
//	if (options.blockSize == 0) {
//		options.blockSize = calculateDefaultBlockSizeForInputFile(inputFile);
//	}
//
//	// TODO - Should we try to extract this from the target URL instead? The case came up when integrating zsync into
//	// Maven deploys where the local POM file is path/pom.xml, but it's uploaded as .../commons-parent-1.0.7.pom, or
//	// something along those lines.
//
//	// filename: default from inputFile
//	if (this.filename == null) {
//		this.filename = inputFile.getFileName().toString();
//	}
//
//	// url: default to filename relative URL
//	if (this.url == null) {
//		this.setUrl(this.filename);
//	}
//
//	return this;
//}
//
//func calculateDefaultBlockSizeForInputFile(fileSize int) uint64 {
//	return fileSize < 100 * 1 << 20 ? BLOCK_SIZE_SMALL : BLOCK_SIZE_LARGE
//}