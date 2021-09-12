package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// See ../../image/README.md for the usage.

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s FILE")
	}

	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}

	fmt.Printf("const %s = \"\" +\n", strings.Replace(args[1], ".", "_", -1))

	i := 0
	max := 32
	for i = 0; i < len(b); i++ {
		bb := b[i]
		if (i % max) == 0 {
			fmt.Printf("	\"")
		}
		fmt.Printf("\\x%02X", bb)
		if (i%max) == max-1 && i != len(b)-1 {
			fmt.Printf("\" + \n")
		}
	}
	if (i % max) < max-1 {
		fmt.Printf("\"\n")
	}

	return nil
}
