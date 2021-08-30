package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

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
