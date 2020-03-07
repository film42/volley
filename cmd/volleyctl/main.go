package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "volleyctl",
		Run: run,
	}
	rootCmd.Flags().String("pid-file", "/tmp/volleyd.pid", "The running volleyd pid file")
	rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalln("Expected command: [start|stop|shutdown]")
	}

	pidFile, _ := cmd.Flags().GetString("pid-file")
	volleydPid, err := getVolleydPid(pidFile)
	if err != nil {
		log.Fatalln("Could not find a volleyd pid file at:", pidFile)
	}
	volleydProcess, err := os.FindProcess(volleydPid)
	if err != nil {
		log.Fatalln("Could not find running volleyd process at pid:", volleydPid)
	}

	switch args[0] {
	case "stop":
		err = volleydProcess.Signal(syscall.SIGUSR1)
	case "start":
		err = volleydProcess.Signal(syscall.SIGUSR2)
	case "shutdown":
		err = volleydProcess.Signal(syscall.SIGABRT)
	default:
		log.Fatalln("Expected command: [start|stop|shutdown], got:", args[0])
	}

	if err != nil {
		log.Fatalln("Error sending signal to volleyd:", err)
	}
}

func getVolleydPid(pidFile string) (int, error) {
	bytes, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return -1, err
	}
	pid, err := strconv.Atoi(string(bytes))
	if err != nil {
		return -1, err
	}
	return pid, nil
}
