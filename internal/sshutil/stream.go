package sshutil

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/melbahja/goph"
	"github.com/relex/aini"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

// linePrinter interleaves output from concurrent hosts one whole line at a
// time, each prefixed with a padded host name.
type linePrinter struct {
	mu    sync.Mutex
	w     io.Writer
	width int
}

func newLinePrinter(w io.Writer, hosts []string) *linePrinter {
	width := 0
	for _, h := range hosts {
		width = max(width, len(h))
	}
	return &linePrinter{w: w, width: width}
}

func (p *linePrinter) print(host, line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.w, "[%s]%s %s\n", green(host), strings.Repeat(" ", p.width-len(host)), line)
}

// streamOnHosts runs the commands on every host at once, printing each output
// line as it arrives. Long-running commands (tail -f, journalctl -f) stream
// until they exit or the user hits Ctrl-C; -timeout and -parallel don't apply.
func streamOnHosts(hosts map[string]*aini.Host, commands []string, opts Options) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	names := slices.Sorted(maps.Keys(hosts))
	printer := newLinePrinter(os.Stdout, names)

	var g errgroup.Group
	errs := make([]error, len(names))
	for i, name := range names {
		g.Go(func() error {
			errs[i] = streamOnHost(ctx, printer, hosts[name], name, commands, opts)
			return nil
		})
	}
	g.Wait()
	return errors.Join(errs...)
}

// streamOnHost connects to one host and streams each command in turn.
func streamOnHost(ctx context.Context, p *linePrinter, h *aini.Host, name string, commands []string, opts Options) error {
	client, err := dialHost(h, opts)
	if err != nil {
		p.print(name, fmt.Sprintf("error: failed to connect: %v", err))
		return fmt.Errorf("%s: %w", name, err)
	}
	defer client.Close()

	var errs []error
	for _, command := range commands {
		if ctx.Err() != nil {
			break
		}
		if err := streamCommand(ctx, p, client, name, command); err != nil {
			p.print(name, fmt.Sprintf("error: %v", err))
			errs = append(errs, fmt.Errorf("%s: %q: %w", name, command, err))
		}
	}
	return errors.Join(errs...)
}

// streamCommand starts the command in its own session and forwards stdout and
// stderr to the printer line by line until the command exits or ctx is
// canceled.
func streamCommand(ctx context.Context, p *linePrinter, client *goph.Client, name, command string) error {
	cmd, err := client.Command(command)
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// goph only applies its context inside Run/Output, not Start/Wait, so
	// cancellation (Ctrl-C) has to tear down the session here.
	unregister := context.AfterFunc(ctx, func() {
		_ = cmd.Signal(ssh.SIGINT)
		_ = cmd.Session.Close()
	})
	defer unregister()

	var wg sync.WaitGroup
	for _, pipe := range []io.Reader{stdout, stderr} {
		wg.Go(func() {
			scanner := bufio.NewScanner(pipe)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for scanner.Scan() {
				p.print(name, scanner.Text())
			}
		})
	}
	wg.Wait()

	err = cmd.Wait()
	if ctx.Err() != nil {
		return nil // interrupted by the user; not a failure
	}
	return err
}
