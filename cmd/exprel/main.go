package main // import "layeh.com/exprel/cmd/exprel"

import (
	"flag"
	"fmt"
	"os"

	"layeh.com/exprel"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [expressions...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\nEvaluates and prints the result of each expression.\n")
	}
	flag.Parse()

	for _, arg := range flag.Args() {
		expr, err := exprel.Parse(arg)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}
		result, err := expr.Evaluate(exprel.Base)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	}
}
