package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"code.google.com/p/go.exp/fsnotify"
)

var mainDir = flag.String("dir", ".", "directory that should be monitored and has pid file in it")
var buildCMD = flag.String("command", "go build .", "command that should be run to build the binary")

// TODO:
// 1. build and run binary before watching
// 2. kill the server on STRG+D/STRG+C
func main() {
	flag.Parse()

	w := New(*mainDir, *buildCMD)
	err := w.Run()
	if err != nil {
		panic(err)
	}
	<-w.Ready
}

func New(dir string, commnd string) (ø *ProjectWatcher) {
	ø = &ProjectWatcher{
		dir:   dir,
		Mutex: &sync.Mutex{},
		Ready: make(chan int, 1),
	}
	ø.buildCMD = commnd

	ø.ReadPID()
	return ø
}

func (p *ProjectWatcher) ReadPID() (pid int) {
	data, err := ioutil.ReadFile(filepath.Join(p.dir, "main.pid"))
	if err != nil {
		panic(err)
	}

	pid, err = strconv.Atoi(string(data))
	if err != nil {
		panic(err)
	}
	return
}

func (p *ProjectWatcher) Reload() {
	c := strings.Split(p.buildCMD, " ")
	cmd := exec.Command(c[0], c[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error while running %#v: %s\n", p.buildCMD, out)
		return
	}

	p.SigUSR2()
}

type ProjectWatcher struct {
	dir string
	*sync.Mutex
	Watcher  *fsnotify.Watcher
	Ready    chan int
	buildCMD string
}

func (ø *ProjectWatcher) Run() error {
	watcher, err := fsnotify.NewWatcher()
	ø.Watcher = watcher
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ev := <-ø.Watcher.Event:
				switch {
				case ev.IsCreate():
					d, err := os.Stat(ev.Name)
					if err == nil {
						if d.IsDir() {
							ø.Lock()
							ø.Watcher.Watch(ev.Name)
							ø.Unlock()
						}
					}
					if filepath.Ext(ev.Name) == ".go" {
						ø.Lock()
						log.Printf("try to reload")
						ø.Reload()
						ø.Unlock()
					}
				case ev.IsDelete():
					d, err := os.Stat(ev.Name)
					if err == nil && (d.IsDir() || filepath.Ext(ev.Name) == ".go") {
						ø.Lock()
						ø.Watcher.RemoveWatch(ev.Name)
						ø.Unlock()
					}

				case ev.IsModify():
				case ev.IsRename():
				}
			case err := <-ø.Watcher.Error:
				log.Println("watcher error:", err)
			}
		}
		ø.Lock()
		ø.Ready <- 1
		ø.Unlock()
	}()

	if err != nil {
		panic("can't create watcher: " + err.Error())
	}
	err = filepath.Walk(ø.dir, filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return filepath.SkipDir
		}
		return ø.Watcher.Watch(path)
	}))

	if err != nil && err != filepath.SkipDir {
		return err
	}

	log.Printf("tea is ready")
	return nil
}

func (ø *ProjectWatcher) SigUSR2() {
	pid := ø.ReadPID()

	cmd := exec.Command("kill", "-s", "USR2", fmt.Sprintf("%v", pid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("can't SIGUSR2 pid %d: %s\n", pid, out)
		return
	}

	log.Printf("reloaded successfully")
}
