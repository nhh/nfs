package pod

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"nfs/internal/config"
	"nfs/internal/helper"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

type ISyncer interface {
	IsRunning() bool
	StartSyncing()
}

type syncer struct {
	id        string
	isRunning bool
	mtx       sync.Mutex
	files     []string
	watchCnf  config.NfsWatchConfig
	podCnf    config.NfsPodConfig
	interval  time.Duration
	watcher   *fsnotify.Watcher
}

func NewSyncer(cnf config.NfsWatchConfig, podConfig config.NfsPodConfig, interval time.Duration) ISyncer {
	return &syncer{id: helper.GenerateId(), watchCnf: cnf, isRunning: false, interval: interval, podCnf: podConfig}
}

func (syncer *syncer) IsRunning() bool {
	return syncer.isRunning
}

func (syncer *syncer) add(file string) {
	syncer.mtx.Lock()
	syncer.files = append(syncer.files, file)
	syncer.mtx.Unlock()
}

func (syncer *syncer) StartSyncing() {
	if syncer.isRunning {
		return
	}

	go syncer.sync()
	go syncer.setupWatcher()
}

// Todo maybe move loop and concurrency settings out of func
func (syncer *syncer) sync() {
	for {
		time.Sleep(syncer.interval)

		if !syncer.isRunning {
			continue
		}

		// Do nothing
		if len(syncer.files) == 0 {
			continue
		}

		syncer.mtx.Lock()

		// Deduplicate
		slices.Sort(syncer.files)
		slices.Compact(syncer.files)

		syncer.files = slices.DeleteFunc(syncer.files, func(i string) bool { return i == "" || strings.HasSuffix(i, "~") })

		selectCommand := fmt.Sprintf("kubectl get pod -n %s -l %s -o name --field-selector 'status.phase==Running'", syncer.podCnf.Namespace, syncer.podCnf.Selector)

		response, err := exec.Command("/bin/bash", "-c", selectCommand).Output()

		if err != nil {
			fmt.Println(err)
			continue
		}

		pods := strings.Split(strings.Trim(string(response), "\n"), " ")
		pods = slices.DeleteFunc(pods, func(i string) bool { return i == "" || strings.HasSuffix(i, "~") })

		for _, pod := range pods {

			start := time.Now()

			fmt.Printf("<%s> %s: Syncing %d files to %s...", syncer.watchCnf.Pattern, start.Format("2006-01-02 15:04:05"), len(syncer.files), pod)

			tmpFile, err := os.CreateTemp("", "kubectl-sync-list")

			// Writing change-list to tmp file
			_, err = tmpFile.Write([]byte(strings.Join(syncer.files, "\n")))

			if err != nil {
				fmt.Printf(" error %v (❌) \n", err)
			}

			cmd := fmt.Sprintf("tar cf - -T %s | kubectl exec -i -n fe-nihanft %s -- tar xf - -C %s", tmpFile.Name(), pod, syncer.podCnf.Cwd)

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
