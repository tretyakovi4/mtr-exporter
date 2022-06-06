package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type mtrJob struct {
	Report   []*mtrReport
	Launched time.Time
	Duration time.Duration

	mtrBinary string
	args      []string
	cmdLine   string

	sync.Mutex
}

func newMtrJob(mtr string, args []string) *mtrJob {
	extra := []string{
		"-j", // json output
	}
	args = append(extra, args...)
	cmd := exec.Command(mtr, args...)

	return &mtrJob{
		args:      args,
		mtrBinary: mtr,
		cmdLine:   strings.Join(cmd.Args, " "),
	}
}

func (job *mtrJob) Launch() error {

	jsonStr, err := ioutil.ReadFile("./url.json")
	domains := make(map[string]interface{})

	json.Unmarshal([]byte(jsonStr), &domains)

	if err != nil {
		panic(err)
	}

	reports := []*mtrReport{}
	launched := time.Now()

	for key, value := range domains {
		args := job.args
		args = append(args, value.(string), key)
		cmd := exec.Command(job.mtrBinary, args...)

		// launch mtr
		buf := bytes.Buffer{}
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			return err
		}

		// decode the report
		report := &mtrReport{}
		if err := report.Decode(&buf); err != nil {
			return err
		}

		reports = append(reports, report)
		// copy the report into the job
	}

	duration := time.Since(launched)

	job.Lock()
	job.Report = reports
	job.Launched = launched
	job.Duration = duration
	job.Unlock()

	// done.
	return nil
}
