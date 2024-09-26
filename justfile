test_file := "test.txt"
evaluate_file := "measurements.txt"

run IMPLEMENTATION_FOLDER FILE:
  go run {{IMPLEMENTATION_FOLDER}}/main.go \
    --file {{FILE}} \
    --cpuprofile {{IMPLEMENTATION_FOLDER}}/cpu.prof \
    --memprofile {{IMPLEMENTATION_FOLDER}}/mem.prof \
    >  {{IMPLEMENTATION_FOLDER}}/result.txt

test IMPLEMENTATION_FOLDER:
  just run {{IMPLEMENTATION_FOLDER}} {{test_file}}

evaluate IMPLEMENTATION_FOLDER:
  just run {{IMPLEMENTATION_FOLDER}} {{evaluate_file}}

cpuprof IMPLEMENTATION_FOLDER:
  go tool pprof {{IMPLEMENTATION_FOLDER}}/cpu.prof

memprof IMPLEMENTATION_FOLDER:
  go tool pprof {{IMPLEMENTATION_FOLDER}}/mem.prof


