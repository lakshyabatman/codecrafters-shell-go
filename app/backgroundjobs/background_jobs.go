package backgroundjobs

import (
	"fmt"
	"io"
)

type ProcessItem struct {
	Command string
	State   string
	Pid     int
}

type BackgroundJobManager struct {
	ProcessList []ProcessItem
}

func (bgJ *BackgroundJobManager) AddProcess(
	pid int,
	stdout io.Writer,
	command string) {

	bgJ.ProcessList = append(bgJ.ProcessList, ProcessItem{
		Command: command,
		State:   "Running",
		Pid:     pid,
	})
	i := len(bgJ.ProcessList)
	fmt.Fprintf(stdout, "[%d] %d\n", i, pid)
}

func (bGj *BackgroundJobManager) List(stdout io.Writer) {
	for idx, process := range bGj.ProcessList {

		// [1]+  Running                 sleep 10 &
		fmt.Fprintf(stdout, "[%d]+  Running                 %s\n", idx+1, process.Command)

	}
}
