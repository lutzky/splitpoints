// splitpoints outputs ffmpeg commands for splitting an input file into a
// series of video files without recoding.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const frameRate = 25

var (
	pointsFile = flag.String("points_file", "points.txt", "Points file, with lines %H:%M:%S:%F-%H:%M:%S:%F")
	inputFile  = flag.String("input_file", "audio_corrected.mov", "Name of input video file")
)

type timecode struct {
	hour   int
	minute int
	second int
	frame  int
}

func (t timecode) frames() int {
	secs := t.second + 60*t.minute + 3600*t.hour
	return frameRate*secs + t.frame
}

func (t timecode) decimal() string {
	msec := 1000 * t.frame / frameRate
	return fmt.Sprintf("%d:%02d:%02d.%03d", t.hour, t.minute, t.second, msec)
}

func parseTimecode(s string) (timecode, error) {
	var r timecode
	_, err := fmt.Sscanf(s, "%d:%02d:%02d:%02d", &r.hour, &r.minute, &r.second, &r.frame)
	if err != nil {
		return r, fmt.Errorf("invalid timecode %q: %w", s, err)
	}
	if r.minute >= 60 || r.second >= 60 || r.frame >= frameRate {
		return r, fmt.Errorf("invalid timecode %q: number too high", s)
	}
	return r, nil
}

func framesToTimecode(frames int) timecode {
	return timecode{
		hour:   frames / (3600 * frameRate),
		minute: (frames / (60 * frameRate)) % 60,
		second: (frames / frameRate) % 60,
		frame:  frames % frameRate,
	}
}

func splitCommand(timecodes, inputFile, outputFile string) (string, error) {
	segments := strings.Split(timecodes, "-")
	if len(segments) != 2 {
		return "", fmt.Errorf("expected START-END")
	}
	start, err := parseTimecode(segments[0])
	if err != nil {
		return "", fmt.Errorf("invalid start time: %w", err)
	}
	end, err := parseTimecode(segments[1])
	if err != nil {
		return "", fmt.Errorf("invalid end time: %w", err)
	}

	length := framesToTimecode(end.frames() - start.frames())

	return fmt.Sprintf("ffmpeg -ss %s -i %s -to %s -c copy %s",
		start.decimal(), inputFile, length.decimal(), outputFile), nil
}

func main() {
	flag.Parse()
	f, err := os.Open(*pointsFile)
	if err != nil {
		log.Fatalf("Error opening %s: %s", *pointsFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		i++
		segmentName := fmt.Sprintf("segment_%d.mov", i)
		cmd, err := splitCommand(scanner.Text(), *inputFile, segmentName)
		if err != nil {
			log.Fatalf("At %s:%d: %s", *pointsFile, i, err)
		}
		fmt.Println(cmd)
	}
}
