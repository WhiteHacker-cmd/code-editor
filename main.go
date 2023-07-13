package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var cli *client.Client

var cntr container.CreateResponse

func homeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "hello world!"}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func createHandler(w http.ResponseWriter, r *http.Request) {

	log.Println(cli)

	containerConfig := &container.Config{
		Image:     "python:alpine",
		OpenStdin: true,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{"c:/Users/007am/Documents/dev/test/:/test/"},
	}

	var err error

	cntr, err = cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		panic(err)
	}

	log.Println(cntr.ID)

	err = cli.ContainerStart(context.Background(), cntr.ID, types.ContainerStartOptions{})

	if err != nil {
		panic(err)
	}

	response := map[string]string{"id": cntr.ID}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func runFileHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"python", "/test/test.py"},
	}

	response, err := cli.ContainerExecCreate(ctx, cntr.ID, execConfig)

	if err != nil {
		panic(err)
	}

	execId := response.ID

	resp, err := cli.ContainerExecAttach(ctx, execId, types.ExecStartCheck{})

	if err != nil {
		panic(err)
	}

	defer resp.Close()

	_, err = io.Copy(os.Stdout, resp.Reader)

	if err != nil {
		panic(err)
	}

	_, err = cli.ContainerExecInspect(ctx, execId)

	if err != nil {
		panic(err)
	}

	responseJson := map[string]string{"message": "hello world!"}

	jsonResponse, err := json.Marshal(responseJson)

	if err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func removeContHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	err := cli.ContainerStop(ctx, cntr.ID, container.StopOptions{})

	if err != nil {
		panic(err)
	}

	err = cli.ContainerRemove(ctx, cntr.ID, types.ContainerRemoveOptions{})

	if err != nil {
		panic(err)
	}

	response := map[string]string{"message": "container removed successfully!"}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func main() {

	var err error
	cli, err = client.NewClient("unix:///var/run/docker.sock", "v1.43", nil, nil)

	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		log.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/run", runFileHandler)
	http.HandleFunc("/remove", removeContHandler)
	log.Println("Server listening on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
