# 1BRC challenge
Solution for [1BRC](https://github.com/gunnarmorling/1brc) (The One Billion Row Challenge)

# Solutions

## Baseline

Simple solution that reads the lines sequentially and aggregate the result line by line (count, min, max and sum).

## Channels

Solution that uses channels to implement MapReduce pattern.

One worker is responsible for reading the file in chunks and sending each chunk to the `jobs` channel.

N consumers retrieve chunks from the `jobs` channel and calculate the statistics (count, min, max and sum of the chunk), sending it to the `result` channel.

The main goroutine consumes the `result` and reduces each chunk result and reports the final statistics.
