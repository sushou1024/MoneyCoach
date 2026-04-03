package app

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
)

const (
	phashSize       = 32
	phashSampleSize = 8
)

func computePHash(data []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	gray := resizeToGray(img, phashSize)
	pixels := make([][]float64, phashSize)
	for y := 0; y < phashSize; y++ {
		row := make([]float64, phashSize)
		for x := 0; x < phashSize; x++ {
			row[x] = float64(gray.GrayAt(x, y).Y)
		}
		pixels[y] = row
	}

	coeffs := dct2D(pixels, phashSampleSize)
	values := make([]float64, 0, phashSampleSize*phashSampleSize-1)
	for y := 0; y < phashSampleSize; y++ {
		for x := 0; x < phashSampleSize; x++ {
			if x == 0 && y == 0 {
				continue
			}
			values = append(values, coeffs[y][x])
		}
	}
	median := medianFloat(values)

	var hash uint64
	bit := 63
	for y := 0; y < phashSampleSize; y++ {
		for x := 0; x < phashSampleSize; x++ {
			if coeffs[y][x] > median {
				hash |= 1 << uint(bit)
			}
			bit--
		}
	}

	return fmt.Sprintf("%016x", hash), nil
}

func resizeToGray(img image.Image, size int) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(image.Rect(0, 0, size, size))
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return gray
	}
	for y := 0; y < size; y++ {
		sy := bounds.Min.Y + y*bounds.Dy()/size
		for x := 0; x < size; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/size
			c := color.GrayModel.Convert(img.At(sx, sy)).(color.Gray)
			gray.SetGray(x, y, c)
		}
	}
	return gray
}

func dct2D(pixels [][]float64, size int) [][]float64 {
	coeffs := make([][]float64, size)
	for i := range coeffs {
		coeffs[i] = make([]float64, size)
	}

	cosX := make([][]float64, size)
	cosY := make([][]float64, size)
	for u := 0; u < size; u++ {
		cosX[u] = make([]float64, phashSize)
		cosY[u] = make([]float64, phashSize)
		for x := 0; x < phashSize; x++ {
			angle := (math.Pi / float64(phashSize)) * (float64(2*x+1) * float64(u) / 2)
			cosX[u][x] = math.Cos(angle)
			cosY[u][x] = math.Cos(angle)
		}
	}

	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			sum := 0.0
			for y := 0; y < phashSize; y++ {
				for x := 0; x < phashSize; x++ {
					sum += pixels[y][x] * cosX[u][x] * cosY[v][y]
				}
			}
			coeffs[v][u] = alpha(u) * alpha(v) * sum
		}
	}
	return coeffs
}

func alpha(k int) float64 {
	if k == 0 {
		return math.Sqrt(1.0 / float64(phashSize))
	}
	return math.Sqrt(2.0 / float64(phashSize))
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func hammingDistanceHex(a, b string) (int, error) {
	if len(a) == 0 || len(b) == 0 {
		return 0, fmt.Errorf("empty hash")
	}
	ab, err := hex.DecodeString(a)
	if err != nil {
		return 0, err
	}
	bb, err := hex.DecodeString(b)
	if err != nil {
		return 0, err
	}
	if len(ab) != len(bb) {
		return 0, fmt.Errorf("hash length mismatch")
	}
	distance := 0
	for i := range ab {
		x := ab[i] ^ bb[i]
		distance += bitsInByte(x)
	}
	return distance, nil
}

func bitsInByte(value byte) int {
	count := 0
	for value > 0 {
		value &= value - 1
		count++
	}
	return count
}
