package mp4util

import (
	"os"
)

// Returns the duration, in seconds, of the mp4 file at the provided filepath
func Duration(filepath string) (int, error) {
	file, _ := os.Open(filepath)
	defer file.Close()

	moovAtomPosition, _, err := findAtom(0, "moov", file)
	if err != nil {
		return 0, err
	}

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

func findAtom(startPos int64, atomName string, file *os.File) (int64, int64, error) {
	buffer := make([]byte, 8)
	for true {
		_, err := file.ReadAt(buffer, startPos)
		if err != nil {
			return 0, 0, err
		}

		lengthOfAtom := int64(convertBytesToInt(buffer[0:4]))

		name := string(buffer[4:])
		if name == atomName {
			return startPos, lengthOfAtom, nil
		}

		startPos += lengthOfAtom
	}
	return -1, 0, nil
}

func durationFromMvhdAtom(mvhdStart int64, mvhdLength int64, file *os.File) (int, error) {
	buffer := make([]byte, 8)
	_, err := file.ReadAt(buffer, mvhdStart+20)
	if err != nil {
		return 0, err
	}

	timescale := convertBytesToInt(buffer[0:4])
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
