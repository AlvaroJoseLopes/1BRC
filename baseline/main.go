package main

import (
	"bufio"
	"flag"
	"fmt"
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


var file = flag.String("file", "", "path to the target file")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

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
            slog.Error("could not create memory profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            slog.Error("could not write memory profile: ", err)
        }
    }

}

func evaluate() {
	fp, err := os.Open(*file)
	if err != nil {
		slog.Error("failed to open file: ", err)
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)

	summaryPerCity := make(map[string]*CityData)

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ";")
		if len(tokens) != 2 {
			slog.Error("unexpected number of tokens after splitting", "n_tokens", len(tokens))
		}

		city := tokens[0]
		temperature, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			slog.Error("error when parsing the temperature", "error", err)
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