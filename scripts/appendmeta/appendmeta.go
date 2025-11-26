package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const find = "<head>"
const append = `
<meta name="go-import" content="gomo.prashantv.com/ git https://github.com/prashantv/go">
`

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify html file to modify")
	}

	run(os.Args[1])
}

func run(filename string) {
	contents := must(os.ReadFile(filename))
	updated := must(appendMeta(string(contents)))
	must1(os.WriteFile(filename, []byte(updated), 0o666))
	fmt.Println("updated", filename)
}

func appendMeta(contents string) (string, error) {
	pre, post, ok := strings.Cut(contents, find)
	if !ok {
		return "", fmt.Errorf("failed to find <head> in html file")
	}

	return pre + find + append + post, nil
}

func must[T any](v T, err error) T {
	must1(err)
	return v
}

func must1(err error) {
	if err != nil {
		panic(err)
	}
}
