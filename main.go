package main

import (
	"sync"
	"time"
)

var (
	startTime = time.Now()
)

func main() {
	gc, err := ReadGithubActionConfig()
	if err != nil {
		Fatalf("%v", err)
	}

	f, err := ReadConfigFile(gc.Config)
	if err != nil {
		Fatalf("%v", err)
	}
	defer f.Close()

	c, err := ParseConfig(gc.Dir, f)
	if err != nil {
		Fatalf("%v", err)
	}

	// Start deploy
	var wg sync.WaitGroup
	wg.Add(len(c.Deploys))
	for _, deploy := range c.Deploys {
		go func() {
			defer wg.Done()
			Run(gc, c, deploy)
		}()
	}
	wg.Wait()
}
