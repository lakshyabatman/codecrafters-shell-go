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

		marker := " "
		if idx == len(bGj.ProcessList)-1 {
			marker = "+"
		} else if idx == len(bGj.ProcessList)-2 {
			marker = "-"
		}
		fmt.Fprintf(stdout, "[%d]%s  %s                 %s\n",
			idx+1, marker, process.State, process.Command)

	}
}
