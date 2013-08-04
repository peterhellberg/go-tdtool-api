package main

import (
	"github.com/bmizerany/pat"
	"io"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	m := pat.New()

	m.Get("/", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			Output(w, exec.Command("tdtool", "-l"))
		}))

	m.Put("/:device/on/sync", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			Output(w, DeviceCommand("--on", req))
		}))

	m.Put("/:device/on", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			Async(w, DeviceCommand("--on", req))
		}))

	m.Put("/:device/off/sync", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			Output(w, DeviceCommand("--off", req))
		}))

	m.Put("/:device/off", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			Async(w, DeviceCommand("--off", req))
		}))

	http.Handle("/", m)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Output(w http.ResponseWriter, cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}
	io.Copy(w, stdout)
	cmd.Wait()
}

func Async(w http.ResponseWriter, cmd *exec.Cmd) {
	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}

	w.WriteHeader(202)
}

func DeviceCommand(option string, req *http.Request) *exec.Cmd {
	device := req.URL.Query().Get(":device")
	return exec.Command("tdtool", option, device)
}
