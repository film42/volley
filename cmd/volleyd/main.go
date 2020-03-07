package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	defer log.Println("Bye")

	runCmd := &cobra.Command{
		Use: "run [command]",
		Run: run,
	}
	rootCmd := &cobra.Command{
		Use: "volleyd",
	}
	rootCmd.AddCommand(runCmd)
	rootCmd.Execute()
}

// If the process stops: EXIT
// If asked to stop: BLOCK
// If asked to start: BLOCK
// If asked to shutdown: STOP and EXIT
func run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalln("Error: You must proive a command to run")
	}
	bin := args[0]

	var binArgs []string
	if len(args) > 1 {
		binArgs = args[1:]
	}

	mgr := &Manager{
		mutex:   sync.Mutex{},
		bin:     bin,
		binArgs: binArgs,
	}
	err := mgr.Start()
	if err != nil {
		log.Fatalln("Error starting the process:", err)
	}
	err = mgr.WaitForUnknownStop()
	if err != nil {
		log.Fatalln("Error from process:", err)
	}
}

type Manager struct {
	mutex             sync.Mutex
	bin               string
	binArgs           []string
	process           *exec.Cmd
	processExitedChan chan error
}

func (m *Manager) Start() error {
	return m.tryStart()
}

func (m *Manager) WaitForUnknownStop() error {
	listenForSignalsChan := m.listenForSignals()
	defer m.cleanupProcess()

	var shouldShutdown bool
	var err error
	for {
		select {
		case sig := <-listenForSignalsChan:
			switch sig {
			case syscall.SIGUSR1:
				// Issue a stop.
				err = m.tryStop()
			case syscall.SIGUSR2:
				// Issue a start.
				err = m.tryStart()
			case syscall.SIGABRT:
				// Issue a shutdown.
				err = m.tryStop()
				shouldShutdown = true
			default:
				// proxy to the process if it's there.
				err = m.trySignalProcess(sig)
			}
		case exitErr := <-m.processExitedChan:
			err = exitErr
			shouldShutdown = true
		}

		if err != nil {
			return err
		}

		if shouldShutdown {
			break
		}
	}
	return nil
}

func (m *Manager) listenForSignals() <-chan os.Signal {
	sigChan := make(chan os.Signal, 100)
	signal.Notify(sigChan,
		// These are for stopping and starting.
		syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGABRT,

		// These are to proxy to the process.
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	return sigChan
}

func (m *Manager) cleanupProcess() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.process != nil {
		m.process.Process.Kill()
		m.process = nil
	}
}

func (m *Manager) trySignalProcess(sig os.Signal) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.process == nil {
		return nil
	}
	return m.process.Process.Signal(sig)
}

func (m *Manager) tryStart() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.process != nil {
		return nil
	}

	log.Println("Starting process:", m.bin, strings.Join(m.binArgs, " "))

	m.process = exec.Command(m.bin, m.binArgs...)
	m.process.Stderr = os.Stderr
	m.process.Stdout = os.Stdout

	// Create the channel waiting on process to finish.
	m.processExitedChan = make(chan error, 1)
	go func() {
		m.processExitedChan <- ignoreSignalErrors(m.process.Run())
	}()
	return nil
}

func (m *Manager) tryStop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.process == nil {
		return nil
	}

	log.Println("Stopping process...")

	// Issue an INT signal.
	err := m.process.Process.Signal(syscall.SIGINT)
	if err != nil {
		return nil
	}

	select {
	case exitErr := <-m.processExitedChan:
		err = exitErr
	case <-time.NewTimer(time.Second * 60).C:
		log.Println("Terminating process...")
		// Issue a TERM signal.
		err = ignoreSignalErrors(m.process.Process.Kill())
	}

	if err != nil {
		return nil
	}
	m.process = nil

	log.Println("Process was stopped.")

	return nil
}

// TODO: Make this more robust.
func ignoreSignalErrors(err error) error {
	switch err.Error() {
	case "signal: interrupt":
		return nil
	case "signal: kill":
		return nil
	case "signal: hangup":
		return nil
	}
	return err
}
