package pod

import (
	"fmt"
	"nfs/internal/config/v1"
	"slices"
	"sync"
	"time"
)

type Syncer interface {
	IsRunning() bool
	Add(file string)
	StartWatching()
}

type syncerImpl struct {
	isRunning bool
	mtx       sync.Mutex
	files     []string
	cnf       v1.NfsConfig
}

func NewSyncer(cnf v1.NfsConfig) Syncer {
	syncer := syncerImpl{files: make([]string, 0), cnf: cnf, isRunning: false}

	return &syncer
}

func (syncer *syncerImpl) IsRunning() bool {
	return syncer.isRunning
}

func (syncer *syncerImpl) Add(file string) {
	syncer.mtx.Lock()
	syncer.files = append(syncer.files, file)
	syncer.mtx.Unlock()
}

func (syncer *syncerImpl) StartWatching() {
	if syncer.isRunning {
		return
	}
	syncer.isRunning = true
	go syncer.watch()
}

// Todo maybe move loop and concurrency settings out of func
func (syncer *syncerImpl) watch() {
	for {
		time.Sleep(time.Duration(syncer.cnf.Interval) * time.Millisecond)

		// Do nothing
		if len(syncer.files) == 0 {
			continue
		}

		syncer.mtx.Lock()

		// Deduplicate
		slices.Sort(syncer.files)
		slices.Compact(syncer.files)

		fmt.Printf("Syncing %s\n", syncer.files)

		// Resetting files to be empty
		syncer.files = syncer.files[:0]

		syncer.mtx.Unlock()
	}
}
