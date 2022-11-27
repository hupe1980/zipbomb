package zipbomb

// Compression methods.
// see APPNOTE.TXT 4.4.5
const (
	Deflate uint16 = 8  // DEFLATE compressed
	BZip2   uint16 = 12 // BZip2 compressed
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

	// Constants for the first byte in CreatorVersion.
	creatorFAT  = 0
	creatorUnix = 3

	// Version numbers.
	// see APPNOTE.TXT 4.4.3.2
	zipVersion20 = 20 // 2.0 - File is compressed using Deflate compression
	zipVersion45 = 45 // 4.5 - File uses ZIP64 format extensions
	zipVersion46 = 46 // 4.6 - File is compressed using BZIP2 compression

	// Limits.
	uint16max = (1 << 16) - 1
	uint32max = (1 << 32) - 1

	// Extra header IDs.
	// See http://mdfs.net/Docs/Comp/Archiving/Zip/ExtraField
	zip64ExtraID = 0x0001 // Zip64 extended information

	IFMT   = 0xf000
	IFSOCK = 0xc000
	IFLNK  = 0xa000
	IFREG  = 0x8000
	IFBLK  = 0x6000
	IFDIR  = 0x4000
	IFCHR  = 0x2000
	IFIFO  = 0x1000
	ISUID  = 0x800
	ISGID  = 0x400
	ISVTX  = 0x200

	msdosDir      = 0x10
	msdosReadOnly = 0x01
)
