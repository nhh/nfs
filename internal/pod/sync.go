package pod

import (
	"fmt"
	"nfs/internal/config/v1"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

func Add(file string) {
	mtx.Lock()
	files = append(files, file)
	mtx.Unlock()
}

var mtx sync.Mutex
var files []string
var cnf = v1.Parse()

func init() {
	go syncWithPod()
}

func syncWithPod() {
	for {
		time.Sleep(time.Duration(cnf.Interval) * time.Millisecond)

		// Do nothing
		if len(files) == 0 {
			continue
		}

		mtx.Lock()

		// Deduplicate
		slices.Sort(files)
		slices.Compact(files)

		for _, watchConfig := range cnf.WatchConfig {
			for _, file := range files {
				isMatching, err := filepath.Match(watchConfig.Pattern, filepath.Base(file))

				if err != nil {
					fmt.Printf("error matching files %s %s\n", watchConfig.Pattern, filepath.Base(file))
					continue
				}

				if !isMatching {
					continue
				}

				fmt.Printf("Syncing %s\n", file)
			}
		}

		// Resetting files to be empty
		files = files[:0]

		mtx.Unlock()
	}
}
