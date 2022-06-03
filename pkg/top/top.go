package top

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type Top struct {
}

func NewTop() *Top {
	return &Top{}
}

//func (q *Queue) State() (*State, error) {
//	return nil, errors.Errorf("TODO")
//}
//
//func (q *Queue) SubmitJob(job *JobRequest) (*JobStatus, error) {
//	return nil, errors.Errorf("TODO -- %+v", job)
//}

func (s *Top) NotFound(w http.ResponseWriter, r *http.Request) {
	logrus.Errorf("HTTPResponder not found from request %+v", r)
	//recordHTTPNotFound(r) // TODO metrics
	http.NotFound(w, r)
}

func (s *Top) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logrus.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	//recordHTTPError(r, err, statusCode) // TODO metrics
	http.Error(w, err.Error(), statusCode)
}
