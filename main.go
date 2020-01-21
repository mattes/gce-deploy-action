package main

import (
	"os"
	"sync"
	"sync/atomic"
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
	var hasErrors uint64
	var wg sync.WaitGroup
	wg.Add(len(c.Deploys))
	for _, deploy := range c.Deploys {
		go func(deploy Deploy) {
			defer wg.Done()

			if err := Run(gc, c, deploy); err != nil {
				atomic.AddUint64(&hasErrors, 1)
				LogError(err.Error(), map[string]string{"name": deploy.Name})
			}

		}(deploy)
	}
	wg.Wait()

	if atomic.LoadUint64(&hasErrors) > 0 {
		os.Exit(1)
	}
}
