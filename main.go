package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const (
	// JPEG markers
	startMarker      = 0xFFD8
	exifMarker       = 0xFFE1
	exifHeader       = "Exif\x00\x00"
	tiffLittleEndian = "II"
	tiffBigEndian    = "MM"
)

func main() {
	//Open the JPEG file
	file, err := os.Open("example.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//Let's read the file
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	data := make([]byte, fileSize)

	_, err = file.Read(data)
	if err != nil {
		log.Fatal("Failed to Read the file", err)
	}

	// Check for Exif data
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		log.Fatal("Invalid JPEG file!")
	}

	//Exif segment (APP1 marker)
	exifOffset := findExifSegment(data)
	if exifOffset == -1 {
		log.Fatal("No Exif data found!")
	}
}

func findExifSegment(data []byte) int {
	for i := 0; i < len(data); i++ {
		if data[i] == 0xFF && data[i+1] == 0xE1 {
			return i + 4 // skip app1 marker length
		}
	}
	return -1
}

func progressExifData(data []byte) {
	// verify the Exif data
	if string(data[:6]) != exifHeader {
		log.Fatal("Invalid Exif header!")
	}

	// check byte algorithm
	var byteOrder binary.ByteOrder

	// check for Little or Big Endian
	byteOrder = binary.BigEndian
	if string(data[6:8]) == tiffLittleEndian {
		byteOrder = binary.LittleEndian
	} else if string(data[6:8]) != tiffBigEndian {
		log.Fatal("Invalid TIFF byte alignment")
	}

	// Read the TIFF header
	offset := int(byteOrder.Uint32(data[10:14]))

	// process the IFD (Image File Directory) to get EXIF data
	processIFD(data[6+offset:], byteOrder)
}

func processIFD(data []byte, byteOrder binary.ByteOrder) {
	numEntries := int(byteOrder.Uint16(data[:2]))

	for i := 0; i < numEntries; i++ {
		entryOffset := 2 + i*12
		tag := byteOrder.Uint16(data[entryOffset : entryOffset+2])
		dataType := byteOrder.Uint16(data[entryOffset+2 : entryOffset+4])
		numValues := byteOrder.Uint32(data[entryOffset+4 : entryOffset+8])
		valueOffset := byteOrder.Uint32(data[entryOffset+8 : entryOffset+12])

		fmt.Printf("Tag: 0x%04X, DataType: %d, NumValues: %d, ValueOffset: %d\n", tag, dataType, numValues, valueOffset)

		switch tag {
		case 0x010F: // Manufacturer tag
			fmt.Println("Manufacturer:", readString(data, valueOffset, numValues))
		case 0x0110: // Camera model tag
			fmt.Println("Camera model:", readString(data, valueOffset, numValues))
		case 0x9003: // Date taken tag
			fmt.Println("Date taken:", readString(data, valueOffset, numValues))
		}
	}
}

func readString(data []byte, offset uint32, length uint32) string {
	return string(data[offset : offset+length])
}
