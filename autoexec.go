package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	flag "github.com/dotcloud/docker/pkg/mflag"
	"gopkg.in/fsnotify.v1"
)

var (
	path    string
	suffix  string
	cmd     string
	verbose bool
	args    []string
	help    bool
)

func init() {
	flag.BoolVar(&help, []string{"h", "#help", "-help"}, false, "display the help")

	flag.BoolVar(&verbose, []string{"v", "#verbose", "-verbose"}, false, "verbose output")

	flag.StringVar(&path, []string{"p", "#path", "-path"}, "", "the path to monitor")
	flag.StringVar(&suffix, []string{"s", "#suffix", "-suffix"}, "", "the file suffix to monitor")

	flag.StringVar(&cmd, []string{"c", "#cmd", "-cmd"}, "", "the exec to run")

	flag.Parse()
}

func startExec(signal chan struct{}, cmd string, args []string) {

	for {
		select {
		case _, ok := <-signal:
			if !ok {
				fmt.Println("Exiting...")
				return
			}
		}

		// chew through other other closely following filesystem events unless it's been quite for 250ms
		var flagged = false
		for {
			select {

			case _, ok := <-signal:
				if !ok {
					fmt.Println("Exiting...")
					return
				}

			case <-time.After(250 * time.Millisecond):
				flagged = true
				break
			}
			if flagged {
				break
			}
		}

		fmt.Println("Running:", cmd, args)

		io.WriteString(os.Stdout, "\033]0;Autoexec - running\007")

		ex := exec.Command(cmd, args...)
		ex.Stdout = os.Stdout
		ex.Stderr = os.Stderr

		err := ex.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = ex.Wait()
		if err != nil {
			fmt.Println(err)
			io.WriteString(os.Stdout, "\033]0;Autoexec - ERROR\007")
		} else {
			io.WriteString(os.Stdout, "\033]0;Autoexec - OK\007")
		}

		fmt.Println()
	}
}

func main() {
	io.WriteString(os.Stdout, "\033]0;Autoexec\007")

	if help || cmd == "" || suffix == "" {
		flag.PrintDefaults()
		return
	}
	fmt.Println(flag.Args())
	args = flag.Args()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if path == "" {

		path, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		fmt.Println("Path: ", path)
	}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error", err)
			return err
		}

		if info.IsDir() && !strings.HasPrefix(path, ".") {
			if verbose {
				fmt.Println("Adding", path)
			}
			watcher.Add(path)
			return nil
		}

		return nil
	})

	defer watcher.Close()

	signal := make(chan struct{})

	// start our worker routine
	go startExec(signal, cmd, args)

	fmt.Println("Monitoring", path, suffix)

	for {
		select {
		case ev := <-watcher.Events:
			if strings.HasSuffix(ev.Name, suffix) {
				if verbose {
					fmt.Printf("[%d] %v\n", time.Now().UnixNano(), ev.Name)
				}
				signal <- struct{}{}
			}

		case err := <-watcher.Errors:
			fmt.Println(err)
		}
	}
}
