package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/bmizerany/pat"
)

func main() {
	m := pat.New()

	m.Get("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Output(w, exec.Command("tdtool", "-l"))
		}))

	m.Put("/:device/on/sync", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Output(w, DeviceCommand("--on", r))
		}))

	m.Put("/:device/on", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Async(w, DeviceCommand("--on", r))
		}))

	m.Put("/:device/off/sync", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Output(w, DeviceCommand("--off", r))
		}))

	m.Put("/:device/off", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Async(w, DeviceCommand("--off", r))
		}))

	http.Handle("/", m)

	port := getenv("PORT", "8080")

	log.Printf("Listening on http://0.0.0.0:%s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Output writes the output of a command to the provided response writer
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

// Async executes a command asynchronously
func Async(w http.ResponseWriter, cmd *exec.Cmd) {
	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}

	w.WriteHeader(202)
}

// DeviceCommand returns a tdtool command for :device and option
func DeviceCommand(option string, r *http.Request) *exec.Cmd {
	return exec.Command("tdtool", option, r.URL.Query().Get(":device"))
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return fallback
}
