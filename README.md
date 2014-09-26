Simple tool to automatically run anything when a watched file changes.

Now includes desktop notifications (tested on Ubuntu) for failures.

Originally written for auto-executing Go builds and tests

### INSTALL

	go get github.com/jboelter/autoexec

### TO RUN

**Building Go Source**

	autoexec -c go -p src -s .go install -v ./...

**Testing Go**

	autoexec -c go -p src -s .go test ./...


**Usage**

	autoexec
	  -c="": the exec to run
	  -h=false: display the help
	  -p="": the path to monitor
	  -s="": the file suffix to monitor
	  -v=false: verbose output


Inspired by github.com/ryanslade/goautotest

### Limitations

There are OS-specific limits as to how many watches can be created.  See the package "gopkg.in/fsnotify.v1"
