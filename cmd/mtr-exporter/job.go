package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	/*
		for key, value := range personMap {
			fmt.Println("index : ", key, " value : ", value)
		}

		domains := []string{
			"us-east-bidder.mathtag.com",
			"33across-us-east.lb.indexww.com",
			"exapi-33across-us-east.rubiconproject.com",
		}
	*/
	reports := []*mtrReport{}

	launched := time.Now()

	for key := range domains {
		args := job.args
		//args = append(args, domains[key])
		args = append(args, key)
		fmt.Println("Key here ", key)
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
