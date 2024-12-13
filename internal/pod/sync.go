package pod

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"nfs/internal/config"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

type Syncer interface {
	IsRunning() bool
	StartSyncing()
}

type syncerImpl struct {
	isRunning bool
	mtx       sync.Mutex
	files     []string
	cnf       config.NfsConfig
	watcher   *fsnotify.Watcher
}

func NewSyncer(cnf config.NfsConfig) Syncer {
	syncer := syncerImpl{files: make([]string, 0), cnf: cnf, isRunning: false}

	return &syncer
}

func (syncer *syncerImpl) IsRunning() bool {
	return syncer.isRunning
}

func (syncer *syncerImpl) add(file string) {
	syncer.mtx.Lock()
	syncer.files = append(syncer.files, file)
	syncer.mtx.Unlock()
}

func (syncer *syncerImpl) StartSyncing() {
	if syncer.isRunning {
		return
	}

	go syncer.sync()
	syncer.isRunning = true

	syncer.setupWatcher()
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

		response, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("kubectl get pod -n %s -l %s -o name --field-selector 'status.phase==Running'", syncer.cnf.PodConfig.Namespace, syncer.cnf.PodConfig.Selector)).Output()

		if err != nil {
			fmt.Println(err)
		}

		pods := strings.Split(strings.Trim(string(response), "\n"), " ")
		pods = slices.DeleteFunc(pods, func(i string) bool { return i == "" || strings.HasSuffix(i, "~") })

		for _, pod := range pods {

			start := time.Now()

			fmt.Printf("%s: Syncing %d files to %s...", start.Format("2006-01-02 15:04:05"), len(syncer.files), pod)

			tmpFile, err := os.CreateTemp("", "kubectl-sync-list")

			// Writing change-list to tmp file
			_, err = tmpFile.Write([]byte(strings.Join(syncer.files, "\n")))

			if err != nil {
				fmt.Printf(" error %v (❌) \n", err)
			}

			cmd := fmt.Sprintf("tar cf - -T %s | kubectl exec -i -n fe-nihanft %s -- tar xf - -C %s", tmpFile.Name(), pod, syncer.cnf.PodConfig.Cwd)

			_, err = exec.Command("/bin/bash", "-c", cmd).Output()

			if err != nil {
				fmt.Printf(" error %v (❌) \n", err)
			}

			err = os.Remove(tmpFile.Name())

			if err != nil {
				fmt.Printf(" error %v (❌) \n", err)
			}

			duration := time.Since(start)

			fmt.Printf(" done in %v (✅)\n", duration)
		}

		// Resetting files to be empty
		syncer.files = syncer.files[:0]

		syncer.mtx.Unlock()
	}
}
