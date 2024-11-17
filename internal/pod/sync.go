package pod

import (
	"fmt"
	"nfs/internal/config/v1"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

type Syncer interface {
	IsRunning() bool
	Add(file string)
	StartSyncing()
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

func (syncer *syncerImpl) StartSyncing() {
	if syncer.isRunning {
		return
	}
	syncer.isRunning = true
	go syncer.sync()
}

// Todo maybe move loop and concurrency settings out of func
func (syncer *syncerImpl) sync() {
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

		syncer.files = slices.DeleteFunc(syncer.files, func(i string) bool { return i == "" || strings.HasSuffix(i, "~") })

		start := time.Now()

		response, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("kubectl get pod -n %s -l %s -o name --field-selector 'status.phase==Running'", syncer.cnf.PodConfig.Namespace, syncer.cnf.PodConfig.Selector)).Output()

		if err != nil {
			fmt.Println(err)
		}

		pods := strings.Split(strings.Trim(string(response), "\n"), " ")
		pods = slices.DeleteFunc(pods, func(i string) bool { return i == "" || strings.HasSuffix(i, "~") })

		for _, pod := range pods {

			cmd := fmt.Sprintf("tar cf - %s | kubectl exec -i -n fe-nihanft %s -- tar xf - -C %s", strings.Join(syncer.files, " "), pod, syncer.cnf.PodConfig.Cwd)

			_, err = exec.Command("/bin/bash", "-c", cmd).Output()

			if err != nil {
				fmt.Println(err)
			}

			duration := time.Since(start)

			fmt.Printf("Synced %s to %s in %v\n", syncer.files, pod, duration)
		}

		// Resetting files to be empty
		syncer.files = syncer.files[:0]

		syncer.mtx.Unlock()
	}
}
