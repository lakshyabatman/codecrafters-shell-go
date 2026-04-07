package backgroundjobs

import (
	"fmt"
	"io"
	"slices"
)

type ProcessItem struct {
	Command string
	State   string
	Pid     int
}

type BackgroundJobManager struct {
	ProcessList  []*ProcessItem
	PidtoProcess map[int]*ProcessItem
}

func (bgJ *BackgroundJobManager) AddProcess(
	pid int,
	stdout io.Writer,
	command string) {
	newPr := ProcessItem{
		Command: command,
		State:   "Running",
		Pid:     pid,
	}
	bgJ.ProcessList = append(bgJ.ProcessList, &newPr)
	bgJ.PidtoProcess[pid] = &newPr
	i := len(bgJ.ProcessList)
	fmt.Fprintf(stdout, "[%d] %d\n", i, pid)
}

func (bGj *BackgroundJobManager) List(stdout io.Writer) {
	garbage := []int{}
	for idx, process := range bGj.ProcessList {
		marker := " "
		if idx == len(bGj.ProcessList)-1 {
			marker = "+"
		} else if idx == len(bGj.ProcessList)-2 {
			marker = "-"
		}
		suffix := "&"
		if process.State == "Done" {
			suffix = ""
		}
		fmt.Fprintf(stdout, "[%d]%s  %s                 %s %s\n",
			idx+1, marker, process.State, process.Command, suffix)
		if process.State == "Done" {
			garbage = append(garbage, process.Pid)
		}
	}

	bGj.ProcessList = slices.DeleteFunc(bGj.ProcessList, func(p *ProcessItem) bool {
		return slices.Contains(garbage, p.Pid)
	})
}

func (bGj *BackgroundJobManager) MarkJobFinished(pid int) {
	bGj.PidtoProcess[pid].State = "Done"
}
