package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.GET("/", handle(Output, exec.Command("tdtool", "-l")))

	router.PUT("/:device/on/sync", handleDevice(Output, "--on"))
	router.PUT("/:device/on", handleDevice(Async, "--on"))
	router.PUT("/:device/off/sync", handleDevice(Output, "--off"))
	router.PUT("/:device/off", handleDevice(Async, "--off"))

	port := getenv("PORT", "8080")

	log.Printf("Listening on http://0.0.0.0:%s", port)
	err := http.ListenAndServe(":"+port, router)
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
func DeviceCommand(option, device string) *exec.Cmd {
	return exec.Command("tdtool", option, device)
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return fallback
}

type executor func(w http.ResponseWriter, cmd *exec.Cmd)

func handle(fn executor, cmd *exec.Cmd) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fn(w, cmd)
	}
}

func handleDevice(fn executor, param string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fn(w, DeviceCommand(param, ps.ByName("device")))
	}
}
