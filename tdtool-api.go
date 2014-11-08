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

	m.Get("/", handle(Output, exec.Command("tdtool", "-l")))

	m.Put("/:device/on/sync", handleDevice(Output, "--on"))
	m.Put("/:device/on", handleDevice(Async, "--on"))
	m.Put("/:device/off/sync", handleDevice(Output, "--off"))
	m.Put("/:device/off", handleDevice(Async, "--off"))

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

type executor func(w http.ResponseWriter, cmd *exec.Cmd)

func handle(fn executor, cmd *exec.Cmd) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(w, cmd)
	})
}

func handleDevice(fn executor, param string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(w, DeviceCommand(param, r))
	})
}
