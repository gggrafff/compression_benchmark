package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"fmt"
	"os"
	"time"

	"github.com/golang/snappy"
	lz4 "github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz/lzma"
)

func main() {
	path := "./test.wasm"
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	compareCompressionCoefficients(data)
	compareCompressionSpeeds(data)
}

func compareCompressionCoefficients(data []byte) {
	fmt.Printf("Original size: %dkb\n", len(data)/1024)

	for _, l := range []int{zlib.HuffmanOnly, 1, 3, 5, 7, 9} {
		zlibData := compressZlib(data, l)
		fmt.Printf("zlib(%d) compressed size: %dkb; compression coefficient: %.2f\n", l, len(zlibData)/1024, float64(len(data))/float64(len(zlibData)))
		if !bytes.Equal(data, decompressZlib(zlibData)) {
			panic("zlib decompression wrong")
		}
	}

	for _, l := range []int{gzip.HuffmanOnly, 1, 3, 5, 7, 9} {
		gzipData := compressGzip(data, l)
		fmt.Printf("gzip(%d) compressed size: %dkb; compression coefficient: %.2f\n", l, len(gzipData)/1024, float64(len(data))/float64(len(gzipData)))
		if !bytes.Equal(data, decompressGzip(gzipData)) {
			panic("gzip decompression wrong")
		}
	}

	for _, o := range []lzw.Order{lzw.MSB, lzw.LSB} {
		for _, lw := range []int{8} {
			lzwData := compressLzw(data, o, lw)
			fmt.Printf("lzw(%d, %d) compressed size: %dkb; compression coefficient: %.2f\n", o, lw, len(lzwData)/1024, float64(len(data))/float64(len(lzwData)))
			if !bytes.Equal(data, decompressLzw(lzwData, o, lw)) {
				panic("lzw decompression wrong")
			}
		}
	}

	for _, l := range []int{flate.HuffmanOnly, 1, 3, 5, 7, 9} {
		flateData := compressFlate(data, l)
		fmt.Printf("flate(%d) compressed size: %dkb; compression coefficient: %.2f\n", l, len(flateData)/1024, float64(len(data))/float64(len(flateData)))
		if !bytes.Equal(data, decompressFlate(flateData)) {
			panic("flate decompression wrong")
		}
	}

	for _, l := range []lz4.CompressionLevel{lz4.Fast, lz4.Level1, lz4.Level3, lz4.Level5, lz4.Level7, lz4.Level9} {
		for _, b := range []lz4.BlockSize{lz4.Block64Kb, lz4.Block256Kb, lz4.Block1Mb, lz4.Block4Mb} {
			lz4Data := compressLz4(data, l, b)
			fmt.Printf("lz4(%d, %d) compressed size: %dkb; compression coefficient: %.2f\n", l, b, len(lz4Data)/1024, float64(len(data))/float64(len(lz4Data)))
			if !bytes.Equal(data, decompressLz4(lz4Data)) {
				panic("lz4 decompression wrong")
			}
		}
	}

	{
		lzmaData := compressLzma(data)
		fmt.Printf("lzma compressed size: %dkb; compression coefficient: %.2f\n", len(lzmaData)/1024, float64(len(data))/float64(len(lzmaData)))
		if !bytes.Equal(data, decompressLzma(lzmaData)) {
			panic("lzma decompression wrong")
		}
	}
	{
		lzma2Data := compressLzma2(data)
		fmt.Printf("lzma2 compressed size: %dkb; compression coefficient: %.2f\n", len(lzma2Data)/1024, float64(len(data))/float64(len(lzma2Data)))
		if !bytes.Equal(data, decompressLzma2(lzma2Data)) {
			panic("lzma2 decompression wrong")
		}
	}
	{
		snappyData := compressSnappy(data)
		fmt.Printf("snappy compressed size: %dkb; compression coefficient: %.2f\n", len(snappyData)/1024, float64(len(data))/float64(len(snappyData)))
		if !bytes.Equal(data, decompressSnappy(snappyData)) {
			panic("snappy decompression wrong")
		}
	}
}

func compareCompressionSpeeds(data []byte) {
	countMeasures := 10

	for _, l := range []int{flate.HuffmanOnly, 1, 3, 5, 7, 9} {
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressZlib(data, l)
		}
		end := time.Now()
		fmt.Printf("zlib(%d) average compression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))

		zlibData := compressZlib(data, l)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressZlib(zlibData)
		}
		end = time.Now()
		fmt.Printf("zlib(%d) average decompression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}

	for _, l := range []int{flate.HuffmanOnly, 1, 3, 5, 7, 9} {
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressGzip(data, l)
		}
		end := time.Now()
		fmt.Printf("gzip(%d) average compression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))

		gzipData := compressGzip(data, l)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressGzip(gzipData)
		}
		end = time.Now()
		fmt.Printf("gzip(%d) average decompression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}

	for _, o := range []lzw.Order{lzw.MSB, lzw.LSB} {
		for _, lw := range []int{8} {
			begin := time.Now()
			for i := 0; i < countMeasures; i++ {
				_ = compressLzw(data, o, lw)
			}
			end := time.Now()
			fmt.Printf("lzw(%d, %d) average compression time: %.2fs\n", o, lw, float64(end.Sub(begin).Seconds())/float64(countMeasures))

			lzwData := compressLzw(data, o, lw)
			begin = time.Now()
			for i := 0; i < countMeasures; i++ {
				_ = decompressLzw(lzwData, o, lw)
			}
			end = time.Now()
			fmt.Printf("lzw(%d, %d) average decompression time: %.2fs\n", o, lw, float64(end.Sub(begin).Seconds())/float64(countMeasures))
		}
	}

	for _, l := range []int{flate.HuffmanOnly, 1, 3, 5, 7, 9} {
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressFlate(data, l)
		}
		end := time.Now()
		fmt.Printf("flate(%d) average compression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))

		flateData := compressFlate(data, l)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressFlate(flateData)
		}
		end = time.Now()
		fmt.Printf("flate(%d) average decompression time: %.2fs\n", l, float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}

	for _, l := range []lz4.CompressionLevel{lz4.Fast, lz4.Level1, lz4.Level3, lz4.Level5, lz4.Level7, lz4.Level9} {
		for _, b := range []lz4.BlockSize{lz4.Block64Kb, lz4.Block256Kb, lz4.Block1Mb, lz4.Block4Mb} {
			begin := time.Now()
			for i := 0; i < countMeasures; i++ {
				_ = compressLz4(data, l, b)
			}
			end := time.Now()
			fmt.Printf("lz4(%d, %d) average compression time: %.2fs\n", l, b, float64(end.Sub(begin).Seconds())/float64(countMeasures))

			lz4Data := compressLz4(data, l, b)
			begin = time.Now()
			for i := 0; i < countMeasures; i++ {
				_ = decompressLz4(lz4Data)
			}
			end = time.Now()
			fmt.Printf("lz4(%d, %d) average decompression time: %.2fs\n", l, b, float64(end.Sub(begin).Seconds())/float64(countMeasures))
		}
	}

	{
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressLzma(data)
		}
		end := time.Now()
		fmt.Printf("lzma average compression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
		lzmaData := compressLzma(data)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressLzma(lzmaData)
		}
		end = time.Now()
		fmt.Printf("lzma average decompression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}

	{
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressLzma2(data)
		}
		end := time.Now()
		fmt.Printf("lzma2 average compression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
		lzma2Data := compressLzma2(data)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressLzma2(lzma2Data)
		}
		end = time.Now()
		fmt.Printf("lzma2 average decompression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}

	{
		begin := time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = compressSnappy(data)
		}
		end := time.Now()
		fmt.Printf("snappy average compression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
		snappyData := compressSnappy(data)
		begin = time.Now()
		for i := 0; i < countMeasures; i++ {
			_ = decompressSnappy(snappyData)
		}
		end = time.Now()
		fmt.Printf("snappy average decompression time: %.2fs\n", float64(end.Sub(begin).Seconds())/float64(countMeasures))
	}
}

func compressZlib(data []byte, level int) []byte {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, level)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressZlib(compressed []byte) []byte {
	var b bytes.Buffer
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	b.ReadFrom(r)
	r.Close()

	return b.Bytes()
}

func compressGzip(data []byte, level int) []byte {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, level)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressGzip(compressed []byte) []byte {
	var b bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	b.ReadFrom(r)
	r.Close()

	return b.Bytes()
}

func compressLzw(data []byte, order lzw.Order, litWidth int) []byte {
	var b bytes.Buffer
	w := lzw.NewWriter(&b, order, litWidth)
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressLzw(compressed []byte, order lzw.Order, litWidth int) []byte {
	var b bytes.Buffer
	r := lzw.NewReader(bytes.NewReader(compressed), order, litWidth)
	b.ReadFrom(r)
	r.Close()

	return b.Bytes()
}

func compressFlate(data []byte, level int) []byte {
	var b bytes.Buffer
	w, err := flate.NewWriter(&b, level)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressFlate(compressed []byte) []byte {
	var b bytes.Buffer
	r := flate.NewReader(bytes.NewReader(compressed))
	b.ReadFrom(r)
	r.Close()

	return b.Bytes()
}

func compressLz4(data []byte, level lz4.CompressionLevel, block lz4.BlockSize) []byte {
	var b bytes.Buffer
	w := lz4.NewWriter(&b)
	if err := w.Apply(lz4.CompressionLevelOption(level), lz4.BlockSizeOption(block)); err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressLz4(compressed []byte) []byte {
	var b bytes.Buffer
	r := lz4.NewReader(bytes.NewReader(compressed))
	b.ReadFrom(r)

	return b.Bytes()
}

func compressLzma(data []byte) []byte {
	var b bytes.Buffer
	w, err := lzma.NewWriter(&b)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressLzma(compressed []byte) []byte {
	var b bytes.Buffer
	r, err := lzma.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	b.ReadFrom(r)

	return b.Bytes()
}

func compressLzma2(data []byte) []byte {
	var b bytes.Buffer
	w, err := lzma.NewWriter2(&b)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressLzma2(compressed []byte) []byte {
	var b bytes.Buffer
	r, err := lzma.NewReader2(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	b.ReadFrom(r)

	return b.Bytes()
}

func compressSnappy(data []byte) []byte {
	var b bytes.Buffer
	w := snappy.NewBufferedWriter(&b)
	w.Write(data)
	w.Close()

	compressed := b.Bytes()
	return compressed
}

func decompressSnappy(compressed []byte) []byte {
	var b bytes.Buffer
	r := snappy.NewReader(bytes.NewReader(compressed))
	b.ReadFrom(r)

	return b.Bytes()
}
