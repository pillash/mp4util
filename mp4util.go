package mp4util

import (
	"fmt"
	"os"
)

func Duration(filepath string) int {
	file, _ := os.Open(filepath)
	defer file.Close()

	moovAtomPosition, _, err := findAtom(0, "moov", file)
	if err != nil {
		fmt.Printf("error finding moov atom: %q\n", err)
		return 0
	}
	//fmt.Printf("moov atom pos: %d\n", moovAtomPosition)

	mvhdAtomPosition, mvhdAtomLength, err := findAtom(moovAtomPosition+8, "mvhd", file)
	if err != nil {
		fmt.Printf("error finding mvhd atom: %q\n", err)
	}
	//fmt.Printf("mvhd atom pos: %d %d\n", mvhdAtomPosition, mvhdAtomLength)

	duration, err := durationFromMvhdAtom(mvhdAtomPosition, mvhdAtomLength, file)
	if err != nil {
		fmt.Printf("error reading duration: %q\n", err)
		return 0
	}
	return duration
}

func findAtom(startPos int64, atomName string, file *os.File) (int64, int64, error) {
	buffer := make([]byte, 8)
	for true {
		_, err := file.ReadAt(buffer, startPos)
		if err != nil {
			return 0, 0, err
		}

		lengthOfAtom := int64(ConvertBytesToInt(buffer[0:4]))
		//fmt.Printf("length of atom: %d\n", lengthOfAtom)

		name := string(buffer[4:])
		//fmt.Printf("atom name: %s\n", atomName)
		if name == atomName {
			//fmt.Printf("found moov atom at %d\n", currentPosition)
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

	timescale := ConvertBytesToInt(buffer[0:4])
	durationInTimeScale := ConvertBytesToInt(buffer[4:])
	return int(durationInTimeScale) / int(timescale), nil
}

func ConvertBytesToInt(buf []byte) int {
	res := 0
	for i := len(buf) - 1; i >= 0; i-- {
		b := int(buf[i])
		shift := uint((len(buf) - 1 - i) * 8)
		res += b << shift
	}
	return res
}

func ConvertSecondsToMinutes(seconds int) (int, int) {
	minutes := seconds / 60
	secsPart := seconds % 60
	return minutes, secsPart
}
