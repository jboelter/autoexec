package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jboelter/notificator"
	fsnotify "gopkg.in/fsnotify.v1"
)

var (
	path    string
	suffix  string
	cmd     string
	verbose bool
	args    []string
	help    bool

	notifier *notificator.Notificator
)

func init() {
	flag.BoolVar(&help, "h", false, "display the help")

	flag.BoolVar(&verbose, "v", false, "verbose output")

	flag.StringVar(&path, "p", "", "the path to monitor")
	flag.StringVar(&suffix, "s", "", "the file suffix to monitor")

	flag.StringVar(&cmd, "c", "", "the exec to run")

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
			text := cmd
			for _, v := range args {
				text += " " + v
			}
			notifier.Push("Autoexec - ERROR", text)
		} else {
			io.WriteString(os.Stdout, "\033]0;Autoexec - OK\007")
		}

		fmt.Println()
	}
}

func main() {
	fmt.Println("Autoexec v1.1")
	io.WriteString(os.Stdout, "\033]0;Autoexec\007")

	if help || cmd == "" || suffix == "" {
		flag.PrintDefaults()
		return
	}
	fmt.Println(flag.Args())
	args = flag.Args()

	notifier = notificator.New(notificator.Options{
		AppName: "Autoexec",
	})

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
			matches, err := filepath.Glob(path + "/*" + suffix)
			if err != nil {
				panic(err)
			}
			if len(matches) > 0 {
				if verbose {
					fmt.Println("Adding  ", path)
				}
				watcher.Add(path)
			} else {
				if verbose {
					fmt.Println("Skipping", path)
				}
			}
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
