package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CityData struct {
	Count 	int
	Min 	float64
	Max 	float64
	Sum 	float64
}

func Chunks(filename string, chunkSize int) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		fp, err := os.Open(*file)
		if err != nil {
			slog.Error("failed to open file", "error", err)
			yield(nil, err)
			return
		}
		defer fp.Close()

		buffer := make([]byte, chunkSize)
		leftover := make([]byte, 0, chunkSize)
		leftoverSize := 0
		for {
			sizeCurrentChunk, err := fp.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				slog.Warn("error reading chunk", "error", err)
			}
			buffer = buffer[:sizeCurrentChunk]
			lastNewLineIndex := bytes.LastIndex(buffer, []byte{'\n'})

			chunk := make([]byte, len(buffer[:lastNewLineIndex]) + leftoverSize)
			chunk = append(leftover, buffer[:lastNewLineIndex]...)

			leftoverSize = len(buffer[lastNewLineIndex+1:])
			leftover = make([]byte, leftoverSize)
			copy(leftover, buffer[lastNewLineIndex+1:])
			
			if !yield(chunk, nil) {
				return
			}
		}
	}
}


var file = flag.String("file", "", "path to the target file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
const chunkSize = 100 * 1024 * 1024 // MB

func main() {
	flag.Parse()

	if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            slog.Error("could not create CPU profile", "error", err)
        }
        defer f.Close() // error handling omitted for example
        if err := pprof.StartCPUProfile(f); err != nil {
            slog.Error("could not start CPU profile", "error", err)
        }
        defer pprof.StopCPUProfile()
    }

	start := time.Now()
	evaluate()
	end := time.Now()
	elapsed := end.Sub(start)
	slog.Info(fmt.Sprintf("elapsed time: %v\n", elapsed))

	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            slog.Error("could not create memory profile", "error", err)
        }
        defer f.Close() // error handling omitted for example
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            slog.Error("could not write memory profile", "error", err)
        }
    }

}

func evaluate() {
	finalCityData := make(map[string]*CityData)
	for chunk, err := range Chunks(*file, chunkSize) {
		if err != nil {
			slog.Error("error getting chunk", "error", err)
			return
		}
		partialData := processChunk(chunk)
		for city, data := range partialData {
			currentCityData, ok := finalCityData[city]
			if !ok {
				finalCityData[city] = data
			} else {
				currentCityData.Count += data.Count
				currentCityData.Sum += data.Sum
				if currentCityData.Max < data.Max {
					currentCityData.Max = data.Max
				}
				if currentCityData.Min > data.Min {
					currentCityData.Min = data.Min
				}
			}
		}
	}

	sortedCities := make([]string, 0, len(finalCityData))
	for key := range finalCityData {
		sortedCities = append(sortedCities, key)
	}
	sort.Strings(sortedCities)

	finalSummary := make([]string, 0, len(sortedCities))
	for _, city := range sortedCities {
		cityData := finalCityData[city]
		mean := cityData.Sum/float64(cityData.Count)
		cityResult := fmt.Sprintf("%s=%.1f/%.1f/%.1f", city, cityData.Min, mean, cityData.Max)
		finalSummary = append(finalSummary, cityResult)
	}
	
	out := fmt.Sprintf("{%s}", strings.Join(finalSummary, ", "))
	fmt.Println(out)
}

func processChunk(chunk []byte) map[string]*CityData {
	summaryPerCity := make(map[string]*CityData)

	chunkString := string(chunk)
	lines := strings.Split(chunkString, "\n")
	for idx, line := range lines {
		tokens := strings.Split(line, ";")
		if len(tokens) != 2 {
			slog.Warn("unexpected number of tokens after splitting","n_tokens", len(tokens), "line", line, "index", idx, "n_lines", len(lines))
			continue
		}

		city := tokens[0]
		temperature, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			slog.Warn("error when parsing the temperature", "error", err)
			continue
		}

		summary, ok := summaryPerCity[city]
		if !ok {
			cityData := CityData{
				Count: 1,
				Min: temperature,
				Max: temperature,
				Sum: temperature,
			}
			summaryPerCity[city] = &cityData
		} else {
			summary.Count += 1
			if temperature < summary.Min {
				summary.Min = temperature
			}

			if temperature > summary.Max {
				summary.Max = temperature
			}

			summary.Sum += temperature
		}
	}

	return summaryPerCity
}