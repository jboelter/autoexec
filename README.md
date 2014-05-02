Simple tool to automatically run anything when a watched file changes.

Originally written for auto-executing Go builds and tests

INSTALL

	go get github.com/joshuaboelter/autoexec

TO RUN

Building Go Source

	autoexec -c go -p src -s .go install -v ./...

Testing Go

	autoexec -c go -p src -s .go test ./...

autoexec
  -c, --cmd="": the exec to run
  -h, --help=false: display the help
  -p, --path="": the path to monitor
  -s, --suffix="": the file suffix to monitor
  -v, --verbose=false: verbose output


Inspired by github.com/ryanslade/goautotest

