package main

import (
	"fmt"
	"time"
)

func main() {
	mph := Build(tokenList)
	start := time.Now()
	for i := 0; i < 100000; i++ {
		for _, token := range tokenList {
			mph.Lookup(token)
		}
	}
	dur := time.Since(start)
	fmt.Println(dur)
}
