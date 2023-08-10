package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

const helpMessage = `Arg Parse

A simple utility to solve the very specific problem of parsing docker build arguments.
Values are parsed in the format <key>='<value>'.
Multiple arguments are seperated by a space.

The default place holder is @@

usage:
	arg-parse [OPTIONS] COMMAND BUILD_ARG ...

flags:
	-p  Placeholder Override
	-v  Verbose Output for debugging

example:
	export BUILD_ARGS="version='v1.1.0' key1='some value' key2='some other value'"
	export BUILD_CMD=$(arg-parse "docker build @@ -t myapp -f Dockerfile ." $BUILD_ARGS)
	# Output:
	# docker build --build-arg="version=v1.1.0" --build-arg="key1=some value" --build-arg="key2=some other value" -t myapp -f Dockerfile .
	eval $BUILD_CMD
`

var defaultPlaceholder string = "@@"

func main() {
	placeholderFlag := flag.String("p", defaultPlaceholder, "placeholder used to replace in the input string")
	verboseFlag := flag.Bool("v", false, "Verbose output for debugging")
	flag.Parse()

	if *verboseFlag {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}})
		slog.SetDefault(slog.New(h))
	}

	if os.Args[1] == "help" {
		os.Stdout.WriteString(helpMessage)
		return
	}

	args := flag.CommandLine.Args()
	commandStr := args[0]
	slog.Debug("Command String: " + commandStr)
	buildArgs := ""
	slog.Debug("place holder flag: " + *placeholderFlag)
	if len(args) > 1 {
		buildArgs = strings.Join(args[1:], " ")
	}
	slog.Debug("Build Args: " + buildArgs)

	m, parseError := parse(strings.NewReader(buildArgs))
	if parseError != nil {
		panic(parseError)
	}

	slog.Debug(fmt.Sprint(m))
	output := strings.ReplaceAll(commandStr, *placeholderFlag, dockerBuildArgStringify(m))

	fmt.Print(output)
}

func dockerBuildArgStringify(m map[string]string) string {
	args := make([]string, len(m))
	for key, value := range m {
		s := fmt.Sprintf("--build-arg=\"%s=%s\"", key, strings.ReplaceAll(value, string('"'), string('\\')+string('"')))
		args = append(args, s)
	}
	return strings.TrimSpace(strings.Join(args, " "))
}

func parse(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanRunes)
	m := make(map[string]string, 0)
	for scanner.Scan() {
		keyName, err := scanKey(scanner)
		if err != nil {
			return m, err
		}
		slog.Debug("New Key: " + keyName)

		value, err := scanValue(scanner)
		if err != nil {
			return m, err
		}
		slog.Debug("New value: " + value)
		m[keyName] = value
	}
	return m, nil
}

func scanKey(scanner *bufio.Scanner) (string, error) {
	keyName := scanner.Text()
	for scanner.Scan() {
		if scanner.Text() == "=" && len(keyName) == 0 {
			return keyName, errors.New("invalid format, reached '=' with key length of 0")
		}
		if scanner.Text() == "=" {
			return strings.TrimSpace(keyName), nil
		}
		keyName += scanner.Text()
	}

	if len(keyName) == 0 {
		// return keyName, errors.New("invalid format, reached '=' with key length of 0")
		return keyName, errors.New("key length of 0")
	}
	return keyName, fmt.Errorf("unexpcted EOF, no value for key '%s'", keyName)
}

func scanValue(scanner *bufio.Scanner) (string, error) {
	valueName := ""
	if !scanner.Scan() {
		return valueName, fmt.Errorf("want: single quote (') opening character, got: EOF")
	}
	if scanner.Text() != "'" {
		return valueName, fmt.Errorf("want: single quote (') opening character, got: '%s'", scanner.Text())
	}

	for scanner.Scan() {
		if scanner.Text() == "'" {
			return valueName, nil
		}
		valueName += scanner.Text()
	}

	return valueName, fmt.Errorf("want: single quote (') closing character, got: EOF")
}
