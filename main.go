package main

import (
	"flag"
	"fmt"
)

var (
	omgiliURL *string
)

func main() {
	omgiliURL = flag.String("url", "http://bitly.com/nuvi-plz", "URL for feed list")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *omgiliURL)
}
