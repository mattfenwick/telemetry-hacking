package queue

import (
	"fmt"
	"github.com/mattfenwick/telemetry-hacking/pkg/utils"
	"io/ioutil"
	"net/http"

	//"github.com/open-telemetry/opentelemetry-go-contrib/"

	log "github.com/sirupsen/logrus"
)

type JobState string

const (
	JobStateTodo       JobState = "JobStateTodo"
	JobStateInProgress JobState = "JobStateInProgress"
	JobStateError      JobState = "JobStateError"
	JobStateSuccess    JobState = "JobStateSuccess"
)

type State struct {
	Finished   []*JobStatus
	InProgress []*JobStatus
}

type JobStatus struct {
	Request *JobRequest
	State   JobState
	Answer  any
	Error   string
}

type JobRequest struct {
	JobId    string
	Function string
	Args     []string
}

type Responder interface {
	State() (*State, error)
	SubmitJob(job *JobRequest) (*JobStatus, error)

	NotFound(w http.ResponseWriter, r *http.Request)
	Error(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}

func SetupHTTPServer(responder Responder) {
	// state of the program
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state, err := responder.State()
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, utils.DumpJSON(state))
		} else {
			responder.NotFound(w, r)
		}
	})

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("unable to read body for AddJob POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			var job JobRequest
			err = utils.ParseJson(&job, body)
			if err != nil {
				log.Errorf("unable to ummarshal JSON for AddJob POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			jobStatus, err := responder.SubmitJob(&job)
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, utils.DumpJSON(jobStatus))
		default:
			responder.NotFound(w, r)
		}
	})
}
