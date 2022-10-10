package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const API_BASE_URL = "https://api.bitbucket.org/2.0"

var workspace, repository, basicAuthKey string

type Target struct {
	RefName string `json:"ref_name"`
}

type State struct {
	Name string `json:"name"`
}

type Pipeline []struct {
	Uuid        string `json:"uuid"`
	BuildNumber int64  `json:"build_number"`
	State       State  `json:"state"`
	Target      Target `json:"target"`
}

type ApiResult struct {
	Page      int      `json:"page"`
	PageLen   int      `json:"pagelen"`
	Pipelines Pipeline `json:"values"`
}

func ListPipelines() ApiResult {
	url := fmt.Sprintf("%s/repositories/%s/%s/pipelines/?page=1&pagelen=10&sort=-created_on", API_BASE_URL, workspace, repository)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicAuthKey))

	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("ListPipelines ERROR. HTTP status %d expected %d", res.StatusCode, http.StatusOK))
	}

	var listPipelines ApiResult

	e := json.Unmarshal(body, &listPipelines)

	if e != nil {
		panic(e)
	}

	return listPipelines
}

func StopPipeline(uuid string) (bool, string) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pipelines/%s/stopPipeline", API_BASE_URL, workspace, repository, uuid)
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return false, err.Error()
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicAuthKey))

	res, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err.Error()
	}

	switch res.StatusCode {
	case http.StatusBadRequest:
		return false, "FAILED - Pipeline already stopped"
	case http.StatusNotFound:
		return false, "FAILED - Pipeline not found"
	case http.StatusNoContent:
		return true, "OK"
	default:
		return false, fmt.Sprintf("FAILED - Undefined status %d", res.StatusCode)
	}
}

func main() {
	if os.Getenv("BITBUCKET_WORKSPACE") == "" ||
		os.Getenv("BITBUCKET_REPO_SLUG") == "" ||
		os.Getenv("BITBUCKET_BRANCH") == "" ||
		os.Getenv("BITBUCKET_PIPELINE_UUID") == "" ||
		os.Getenv("BITBUCKET_BUILD_NUMBER") == "" ||
		os.Getenv("TDP_BITBUCKET_BASIC_AUTH") == "" {
		panic("Missing required env variables")
	}

	workspace = os.Getenv("BITBUCKET_WORKSPACE")
	repository = os.Getenv("BITBUCKET_REPO_SLUG")
	basicAuthKey = os.Getenv("TDP_BITBUCKET_BASIC_AUTH")

	branchName := os.Getenv("BITBUCKET_BRANCH")
	currentPipelineUuid := os.Getenv("BITBUCKET_PIPELINE_UUID")
	currentBuildNumber, _ := strconv.ParseInt(os.Getenv("BITBUCKET_BUILD_NUMBER"), 10, 64)

	fmt.Println("Looking for duplicate " + branchName + " pipelines to terminate")

	r := ListPipelines()

	hasTerminated := false
	for _, pipeline := range r.Pipelines {
		if pipeline.Target.RefName == branchName &&
			pipeline.State.Name != "COMPLETED" &&
			pipeline.Uuid != currentPipelineUuid &&
			pipeline.BuildNumber < currentBuildNumber {
			fmt.Printf(". Terminating build number %d, uuid \"%s\"\t", pipeline.BuildNumber, pipeline.Uuid)

			_, pipelineStoppedMessage := StopPipeline(pipeline.Uuid)

			fmt.Println(pipelineStoppedMessage)

			hasTerminated = true
		}
	}

	if !hasTerminated {
		fmt.Println("All is well, no duplicate pipelines found!")
	} else {
		fmt.Println("Done!")
	}
}
