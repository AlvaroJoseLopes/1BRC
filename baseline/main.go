package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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


var file = flag.String("file", "", "path to the target file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

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
	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)

	summaryPerCity := make(map[string]*CityData)

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ";")
		if len(tokens) != 2 {
			log.Fatal("unexpected number of tokens after splitting: ", len(tokens))
		}

		city := tokens[0]
		temperature, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			log.Fatal("error when parsing the temperature: ", err)
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

	sortedCities := make([]string, 0, len(summaryPerCity))
	for key := range summaryPerCity {
		sortedCities = append(sortedCities, key)
	}
	sort.Strings(sortedCities)

	result := make([]string, 0, len(sortedCities))
	for _, city := range sortedCities {
		cityData := summaryPerCity[city]
		mean := cityData.Sum/float64(cityData.Count)
		cityResult := fmt.Sprintf("%s=%.1f/%.1f/%.1f", city, cityData.Min, mean, cityData.Max)
		result = append(result, cityResult)
	}
	
	finalSummary := fmt.Sprintf("{%s}", strings.Join(result, ", "))
	fmt.Println(finalSummary)
}