package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type status struct {
	Gitlab   int           `json:"gitlab"`
	Database int           `json:"database"`
	Uptime   time.Duration `json:"uptime"`
	Version  string        `json:"version"`
}

type userIssue struct {
	Users []user `json:"users"`
	Auth  bool   `json:"auth"`
}

type user struct {
	Username string `json:"username"`
	Count    int    `json:"count"`
}

type labelIssue struct {
	Labels []label `json:"labels"`
	Auth   bool    `json:"auth"`
}

type label struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type author struct {
	Author username `json:"author"`
}

type username struct {
	Username string `json:"username"`
}

type project struct {
	ID   int    `json:"id"`
	Path string `json:"path_with_namespace"`
}

type incomingProject struct {
	Project string `json:"project"`
}

var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

func issueHandler(w http.ResponseWriter, r *http.Request) {
	types, ok := r.URL.Query()["type"]

	if !ok || len(types[0]) < 1 {
		fmt.Fprint(w, "URL param 'type' is missing")
		return
	}

	t := types[0]
	var incProj incomingProject
	err := json.NewDecoder(r.Body).Decode(&incProj)
	if err != nil {
		http.Error(w, "No project", http.StatusInternalServerError)
	}

	var projects []project
	resp, err2 := http.Get("https://git.gvk.idi.ntnu.no/api/v4/projects")
	if err2 != nil {
		http.Error(w, "No project", http.StatusInternalServerError)
	}

	json.NewDecoder(resp.Body).Decode(&projects)

	var projectID int
	for _, v := range projects {
		if v.Path == incProj.Project {
			projectID = v.ID
			break
		}
	}

	if string(t) == "users" {
		var authors []author
		resp, err = http.Get("https://git.gvk.idi.ntnu.no/api/v4/projects/" + string(projectID) + "/issues")
		if err != nil {
			http.Error(w, "Could not find GitLab", http.StatusServiceUnavailable)
		}

		json.NewDecoder(resp.Body).Decode(&authors)
		var m map[string]int
		for _, v := range authors {
			_, ok := m[v.Author.Username]
			if ok {
				m[v.Author.Username]++
			} else {
				m[v.Author.Username] = 1
			}
		}

		var userIss userIssue
		for k, v := range m {
			userIss.Users[v].Username = k
			userIss.Users[v].Count = m[k]
		}

		userIss.Auth = false

		json.NewEncoder(w).Encode(userIss)
	} else if string(t) == "labels" {
		//when type == labels
	}

}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	var stat status

	resp, err := http.Get("git.gvk.idi.ntnu.no/api/v4/projects")
	stat.Gitlab = resp.StatusCode
	if err != nil {
		http.Error(w, "Could not find GitLab", http.StatusServiceUnavailable)
	}

	stat.Uptime = uptime() / 1000000000

	stat.Version = "v1"

}

func main() {

}
