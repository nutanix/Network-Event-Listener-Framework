// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Main package for the event consumer service.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
)

const (
  // PaloAlto Config Directory Path
  LogsDir = "/opt/pafw/logs/"
  BinDir = "/opt/pafw/bin"
  EventConsumer = "pafweventconsumer"
  PidRoot = "/proc" // Includes a directory for each running process.
)


// Usage for command line parameters
func usage() {
	fmt.Fprintf(os.Stderr, "usage: eventconsumer [start|stop|restart|status]\n", )
	os.Exit(2)
}

//Validates service command line options.
//
// Input:
//	service: Command line input string.
// Output:
//	Error if options are not valid.
//
func validateInputParams(service string) error {
	if !(service == "start" || service == "stop" || service == "restart" || service == "status") {
		return errors.New("Not valid service command.")
	}
	return nil
}

// Create directory if not exists
func createDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0755)
	} else {
		return err
	}
	return nil
}

// Initialization function.
// Validate Input parameters.
// if provided incorrect input command, exits with usage help.
func init() {
	if(len(os.Args) > 1){
		if err := validateInputParams(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "!! Error:- %s\n", err.Error())
			usage()
		} else {
			if err = createDirectory(LogsDir); err != nil {
				fmt.Fprintf(os.Stderr, "!! Failed to create logs dir. Error:- %s\n", err.Error())
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "!! Error:- Not sufficient arguments...\n", )
		usage()
	}
}

func findProcess(processName string, pid *int) filepath.WalkFunc {
	return func(path string, _ os.FileInfo, err error) error {

	    // We are only interested in files with a path looking like /proc/<pid>/status.
	    if strings.Count(path, "/") == 3 && strings.Contains(path, "/status") {

		    // Let's extract the middle part of the path with the <pid> and
		    // convert the <pid> into an integer. Log an error if it fails.
		    processId, err := strconv.Atoi(path[6:strings.LastIndex(path, "/")])
		    if err != nil {
			fmt.Println(err)
			return nil
		    }
		    // The status file contains the name of the process in its first line.
		    // The line looks like "Name: theProcess".
		    // Log an error in case we cant read the file.
		    f, err := ioutil.ReadFile(path)
		    if err != nil {
			fmt.Println(err)
			return nil
		    }
		    // Extract the process name from within the first line in the buffer
		    name := string(f[6:bytes.IndexByte(f, '\n')])
		    if(name == processName[0:15]) {
			    *pid = processId
		    }
		}
	    return nil
	}
}

// Checks if Event consumer is already running.
// Return:
// 	True if Event Consumer is running else False
func isEventConsumerRunning(processName string) (bool, int, error) {
	pid := 0
	err := filepath.Walk(PidRoot, findProcess(processName, &pid))
	if pid != 0 {
		return true, pid, nil
	} else {
		if(err != nil) {
			fmt.Println(err)
		}
	}
	return false, pid, err
}
// Starts PANFW Event Consumer if not already any instance is running.
func startEventConsumer() error {
	fmt.Fprintf(os.Stdout, "\nStarting %s....", EventConsumer)
	eventConsumer := fmt.Sprintf("%s/%s", BinDir, EventConsumer)
	cmd := exec.Command(eventConsumer, "-stderrthreshold=ERROR", fmt.Sprintf("-log_dir=%s", LogsDir))
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if(err != nil) {
		fmt.Fprintf(os.Stderr, "\n!! Error:- %s\n", err.Error())
		return err
	} else {
		fmt.Fprintf(os.Stdout, ".Started successfully !!\n")
	}
	return nil
}

// Kill OS Process.
func killProcess(p *os.Process) error {
	err := p.Kill()
	if(err != nil) {
		return err
	}
	return nil
}

// Stops PANFW Event Consumer if any instance is running.
func stopEventConsumer(pid int) error {
	fmt.Fprintf(os.Stdout, "\nStopping %s....", EventConsumer)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
		fmt.Println(err)
	}
	err = killProcess(proc)
	if(err != nil) {
		fmt.Fprintf(os.Stderr, err.Error())
		return err
	}
	fmt.Fprintf(os.Stdout, ".Stopped\n")
	return nil
}

// Kick start function for handling PAN FW event consumer service.
func main() {
	service:= os.Args[1]
	processRunning, pid, _ := isEventConsumerRunning(EventConsumer)
	if(service == "start") {
		if(processRunning == true) {
			fmt.Fprintf(os.Stdout, "Process %s is already running with pid %d. You may want to kill or restart process...\n", EventConsumer, pid)
		} else {
			startEventConsumer()
		}
	}
	if(service == "stop") {
		if(processRunning == true) {
			stopEventConsumer(pid)
		} else {
			fmt.Fprintf(os.Stdout, " !! No instance of '%s' is running.", EventConsumer)
		}
	}
	if(service == "restart") {
		if(processRunning == true) {
			stopEventConsumer(pid)
		}
		startEventConsumer()
	}
	if(service == "status") {
		if(processRunning == true) {
			fmt.Fprintf(os.Stdout, "Process %s is running with pid %d.\n", EventConsumer, pid)
		} else {
			fmt.Fprintf(os.Stdout, "No instance of '%s' is running..", EventConsumer)
		}
	}
	fmt.Println("\n")
}
