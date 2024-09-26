package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CityData struct {
	Count 	int
	Min 	float64
	Max 	float64
	Sum 	float64
}


var file = flag.String("file", "", "path to the target file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
const chunkSize = 1 * 1024 * 1024 // 1MB

func main() {
	flag.Parse()

	if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }

	start := time.Now()
	evaluate()
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("elapsed time: %v\n", elapsed)

	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
    }

}

func evaluate() {
	fp, err := os.Open(*file)
	if err != nil {
		log.Fatal("failed to open file: ", err)
	}
	defer fp.Close()

	jobs := make(chan []byte, 10)
	result := make(chan map[string]*CityData, 7)
	var wg sync.WaitGroup

	// spawn worker goroutines that gets the summary for each chunk 
	for i := 0; i < runtime.NumCPU() - 1; i++ {
		wg.Add(1)
		go func ()  {
			for chunk := range jobs {
				processChunk(chunk, result)
			}
			wg.Done()
		}()
	}


	// spawn goroutine to read file in chunks and send each chunk to jobs channel
	go func ()  {
		buffer := make([]byte, chunkSize)
		leftover := make([]byte, 0, chunkSize)
		leftoverSize := 0
		for {
			sizeCurrentChunk, err := fp.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("error reading chunk: %v", err)
			}
			buffer = buffer[:sizeCurrentChunk]
			lastNewLineIndex := bytes.LastIndex(buffer, []byte{'\n'})

			job := make([]byte, len(buffer[:lastNewLineIndex+1]) + leftoverSize)
			job = append(leftover, buffer[:lastNewLineIndex+1]...)

			leftoverSize = len(buffer[lastNewLineIndex+1:])
			leftover = make([]byte, leftoverSize)
			copy(leftover, buffer[lastNewLineIndex+1:])
			
			jobs <- job
		}
		close(jobs)
		wg.Wait()
		close(result)
	}()

	finalCityData := make(map[string]*CityData)
	for partialData := range result {
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

func processChunk(chunk []byte, result chan<- map[string]*CityData) {
	summaryPerCity := make(map[string]*CityData)

	chunkString := string(chunk)
	for _, line := range strings.Split(chunkString, "\n") {
		tokens := strings.Split(line, ";")
		if len(tokens) != 2 {
			log.Print("unexpected number of tokens after splitting: ", len(tokens), " line: \"", line, "\"")
			continue
		}

		city := tokens[0]
		temperature, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			log.Print("error when parsing the temperature: ", err)
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

	result <- summaryPerCity
}