package overlap

// Compression methods.
const (
	Deflate uint16 = 8 // DEFLATE compressed
)

const (
	fileHeaderSignature      = 0x04034b50
	directoryHeaderSignature = 0x02014b50
	directoryEndSignature    = 0x06054b50
	directory64LocSignature  = 0x07064b50
	directory64EndSignature  = 0x06064b50
	fileHeaderLen            = 30 // + filename + extra
	directoryHeaderLen       = 46 // + filename + extra + comment
	directoryEndLen          = 22 // + comment
	directory64LocLen        = 20 //
	directory64EndLen        = 56 // + extra

	// Version numbers.
	zipVersion20 = 20 // 2.0
	zipVersion45 = 45 // 4.5 (reads and writes zip64 archives)

	// Limits.
	uint16max = (1 << 16) - 1
	uint32max = (1 << 32) - 1

	// Extra header IDs.
	// See http://mdfs.net/Docs/Comp/Archiving/Zip/ExtraField
	zip64ExtraID = 0x0001 // Zip64 extended information
)
