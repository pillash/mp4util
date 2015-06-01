package mp4util

import (
	"os"
)

// Returns the duration, in seconds, of the mp4 file at the provided filepath.
// If an error occurs, the error returned is non-nil
func Duration(filepath string) (int, error) {
	file, _ := os.Open(filepath)
	defer file.Close()

	moovAtomPosition, _, err := findAtom(0, "moov", file)
	if err != nil {
		return 0, err
	}

	// start searching for the mvhd atom inside the moov atom.
	// The first child atom of the moov atom starts 8 bytes after the start of the moov atom.
	mvhdAtomPosition, mvhdAtomLength, err := findAtom(moovAtomPosition+8, "mvhd", file)
	if err != nil {
		return 0, err
	}

	duration, err := durationFromMvhdAtom(mvhdAtomPosition, mvhdAtomLength, file)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// Finds the starting position of the atom of the given name if it is a direct child of the atom
// that is indicated by the given start position.
// Returns: If found the starting byte position of atom is returned along with the atom's size.
//          If not found, -1 is returned as the starting byte position
//          If there was an error, the error is non-nil
func findAtom(startPos int64, atomName string, file *os.File) (int64, int64, error) {
	buffer := make([]byte, 8)
	for true {
		_, err := file.ReadAt(buffer, startPos)
		if err != nil {
			return 0, 0, err
		}

		// The structure of an mp4 atom is:
		// 4 bytes - length of atom
		// 4 bytes - name of atom in ascii encoding
		// rest    - atom data
		lengthOfAtom := int64(convertBytesToInt(buffer[0:4]))
		name := string(buffer[4:])
		if name == atomName {
			return startPos, lengthOfAtom, nil
		}

		// move to next atom's starting position
		startPos += lengthOfAtom
	}
	return -1, 0, nil
}

// Returns the duration in seconds as given by the data in the mvhd atom starting at mvhdStart
// Returns non-nill error is there is an error.
func durationFromMvhdAtom(mvhdStart int64, mvhdLength int64, file *os.File) (int, error) {
	buffer := make([]byte, 8)
	_, err := file.ReadAt(buffer, mvhdStart+20) // The timescale field starts at the 21st byte of the mvhd atom
	if err != nil {
		return 0, err
	}

	// The timescale is bytes 21-24.
	// The duration is bytes 25-28
	timescale := convertBytesToInt(buffer[0:4]) // This is in number of units per second
	durationInTimeScale := convertBytesToInt(buffer[4:])
	return int(durationInTimeScale) / int(timescale), nil
}

func convertBytesToInt(buf []byte) int {
	res := 0
	for i := len(buf) - 1; i >= 0; i-- {
		b := int(buf[i])
		shift := uint((len(buf) - 1 - i) * 8)
		res += b << shift
	}
	return res
}
