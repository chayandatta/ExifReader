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
	startMarker      = 0xFFD8 // JPEG start of image marker
	exifMarker       = 0xFFE1 // APP1 marker where EXIF data is stored
	exifHeader       = "Exif\x00\x00"
	tiffLittleEndian = "II" // Intel TIFF header
	tiffBigEndian    = "MM" // Motorola TIFF header
)

func main() {
	// Open the JPEG file
	file, err := os.Open("images/example.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the entire file into memory
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	data := make([]byte, fileSize)

	_, err = file.Read(data)
	if err != nil {
		log.Fatal("Failed to read file:", err)
	}

	// Check if the file is a JPEG and contains EXIF data
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		log.Fatal("Not a valid JPEG file")
	}

	// Search for the EXIF segment (APP1 marker)
	exifOffset := findEXIFSegment(data)
	if exifOffset == -1 {
		log.Fatal("No EXIF data found")
	}

	// Process the EXIF header and data
	processEXIFData(data[exifOffset:])
}

func findEXIFSegment(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xE1 {
			return i + 4 // Skip the APP1 marker and length
		}
	}
	return -1
}

func processEXIFData(data []byte) {
	// Verify EXIF header
	if string(data[:6]) != exifHeader {
		log.Fatal("Invalid EXIF header")
	}

	// check for byte algorithm
	var byteOrder binary.ByteOrder
	// Check byte alignment (Little or Big Endian)
	byteOrder = binary.BigEndian
	if string(data[6:8]) == tiffLittleEndian {
		byteOrder = binary.LittleEndian
	} else if string(data[6:8]) != tiffBigEndian {
		log.Fatal("Invalid TIFF byte alignment")
	}

	// Read the TIFF header
	offset := int(byteOrder.Uint32(data[10:14]))

	// Process the IFD (Image File Directory) to get EXIF data
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
			fmt.Println("Camera Model:", readString(data, valueOffset, numValues))
		case 0x9003: // Date taken tag
			fmt.Println("Date Taken:", readString(data, valueOffset, numValues))
		}
	}
}

func readString(data []byte, offset uint32, length uint32) string {
	return string(data[offset : offset+length])
}
