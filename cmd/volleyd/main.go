package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	runCmd := &cobra.Command{
		Use: "run [command]",
		Run: run,
	}
	runCmd.Flags().String("pid-file", "/tmp/volleyd.pid", "File to write the volleyd pid while running")
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

	pidFile, _ := cmd.Flags().GetString("pid-file")
	if pidFileExists(pidFile) {
		log.Fatalf("A pid file named %s already exists. Is another volleyd process running?", pidFile)
	}
	createPidFile(pidFile)
	defer deletePidFile(pidFile)

	mgr := &Manager{
		mutex:   sync.Mutex{},
		bin:     bin,
		binArgs: binArgs,
	}
	err := mgr.Start()
	if err != nil {
		// Required because Fatalln exits before defer fires.
		deletePidFile(pidFile)
		log.Fatalln("Error starting the process:", err)
	}
	err = mgr.WaitForUnknownStop()
	if err != nil {
		// Required because Fatalln exits before defer fires.
		deletePidFile(pidFile)
		log.Fatalln("Error from process:", err)
	}
}

func pidFileExists(pidFile string) bool {
	info, err := os.Stat(pidFile)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createPidFile(pidFile string) error {
	return ioutil.WriteFile("/tmp/volleyd.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
}

func deletePidFile(pidFile string) {
	err := os.Remove(pidFile)
	if err != nil {
		log.Println("Unable to remove pid file:", err)
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
				// Proxy to the process if it's there.
				// If we get a signal and the process is stopped, we will
				// get a shouldShutdown == true and will shutdown because
				// we're being asked to stop.
				err, shouldShutdown = m.trySignalProcess(sig)
			}
		case exitErr := <-m.processExitedChan:
			err = exitErr
			shouldShutdown = true
		}

		if err != nil {
			return err
		}

		if shouldShutdown {
			log.Println("Shutting down...")
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

func (m *Manager) trySignalProcess(sig os.Signal) (error, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.process == nil {
		return nil, true
	}
	return m.process.Process.Signal(sig), false
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

	// Issue a TERM signal.
	err := m.process.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return nil
	}

	select {
	case exitErr := <-m.processExitedChan:
		err = exitErr
		m.process = nil
	case <-time.NewTimer(time.Second * 5).C:
		log.Println("Terminating process...")
		// Issue a KILL signal.
		err = m.process.Process.Signal(syscall.SIGKILL)
	}
	if err != nil {
		return nil
	}

	// If we killed but haven't seen the process exit, wait.
	if m.process != nil {
		err = <-m.processExitedChan
		m.process = nil
	}

	log.Println("Process was stopped.")
	return err
}

// TODO: Make this more robust.
func ignoreSignalErrors(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "signal: interrupt":
		return nil
	case "signal: killed":
		return nil
	case "signal: hangup":
		return nil
	}
	return err
}
