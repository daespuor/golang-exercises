package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type Processor interface {
	countBytes() int
	countLines() int
	countWords() int
	countChars() int
}

type FileProcessor struct {
	filepath string
}

type ContentProcessor struct {
	content []byte
}

func (fp FileProcessor) countBytes() int {
	content, err := os.ReadFile(fp.filepath)
	if err != nil {
		panic("File not found")
	}
	return len(content)
}

func (cp ContentProcessor) countBytes() int {
	return len(cp.content)
}

func (fp FileProcessor) countLines() int {
	var count int
	file, err := os.Open(fp.filepath)
	if err != nil {
		panic("File not found")
	}
	defer func() {
		if err = file.Close(); err != nil {
			panic("Error closing the file")
		}
	}()

	s := bufio.NewScanner(file)

	for s.Scan() {
		count++
	}
	if err = s.Err(); err != nil {
		panic("Error scanning the file")
	}
	return count
}

func (cp ContentProcessor) countLines() int {
	lines := strings.Split(string(cp.content), "\n")

	return len(lines)
}

func (fp FileProcessor) countWords() int {
	var count int
	file, err := os.Open(fp.filepath)
	if err != nil {
		panic("File not found")
	}
	defer func() {
		if err = file.Close(); err != nil {
			panic("Error closing the file")
		}
	}()

	s := bufio.NewScanner(file)

	for s.Scan() {
		line := s.Text()
		words := strings.Fields(line)
		count += len(words)
	}
	if err = s.Err(); err != nil {
		panic("Error scanning the file")
	}
	return count
}

func (cp ContentProcessor) countWords() int {
	words := strings.Fields(string(cp.content))
	return len(words)
}

func (fp FileProcessor) countChars() int {
	content, err := os.ReadFile(fp.filepath)
	if err != nil {
		panic("File not found")
	}
	chars := string(content)
	return len([]rune(chars))
}

func (cp ContentProcessor) countChars() int {
	chars := string(cp.content)
	return len([]rune(chars))
}

func main() {
	var cFlag, lFlag, wFlag, mFlag bool
	var orderedFlags []string
	var processor Processor
	flagSet := flag.NewFlagSet("fs", flag.ContinueOnError)
	flagSet.BoolVar(&cFlag, "c", false, "Count the amount of bytes of a file")
	flagSet.BoolVar(&lFlag, "l", false, "Count the amount of lines of a file")
	flagSet.BoolVar(&wFlag, "w", false, "Count the amount of words of a file")
	flagSet.BoolVar(&mFlag, "m", false, "Count the amount of chars of a file")
	flagSet.Parse(os.Args[1:])

	// Populating orderedFlags
	for _, f := range os.Args[1:] {
		switch f {
		case "-c":
			if cFlag {
				orderedFlags = append(orderedFlags, "c")
			}
		case "-l":
			if lFlag {
				orderedFlags = append(orderedFlags, "l")
			}
		case "-m":
			if mFlag {
				orderedFlags = append(orderedFlags, "m")
			}
		case "-w":
			if wFlag {
				orderedFlags = append(orderedFlags, "w")
			}
		}
	}

	// Get filepath
	otherArgs := flagSet.Args()
	for _, v := range otherArgs {
		processor = FileProcessor{filepath: v}
		break
	}

	// If no filepath get content from stdin
	if processor == nil {
		stat, err := os.Stdin.Stat()

		if err != nil {
			panic("not able to read from stdin")
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				panic("not able to read from stdin")
			}
			processor = ContentProcessor{content: content}
		}
	}

	if processor == nil {
		panic("not processor detected")
	}

	for _, f := range orderedFlags {
		switch f {
		case "c":
			result := processor.countBytes()
			fmt.Printf("%d ", result)
		case "l":
			result := processor.countLines()
			fmt.Printf("%d ", result)
		case "m":
			result := processor.countChars()
			fmt.Printf("%d ", result)
		case "w":
			result := processor.countWords()
			fmt.Printf("%d ", result)
		}
	}

	if len(orderedFlags) == 0 {
		fmt.Printf("%d %d %d ", processor.countBytes(), processor.countLines(), processor.countWords())
	}

	processorType := reflect.TypeOf(processor)
	processorValue := reflect.ValueOf(processor)

	if processorType.Name() == "FileProcessor" {
		value := processorValue.FieldByName("filepath")
		fmt.Println(value)
	}

}
