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

	router.GET("/", handle(Output, []string{"tdtool", "-l"}))

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
func Output(w http.ResponseWriter, args []string) {
	cmd := exec.Command("tdtool", args...)

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
func Async(w http.ResponseWriter, args []string) {
	cmd := exec.Command("tdtool", args...)

	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}

	w.WriteHeader(202)
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return fallback
}

type executor func(w http.ResponseWriter, args []string)

func handle(fn executor, args []string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fn(w, args)
	}
}

func handleDevice(fn executor, param string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fn(w, []string{param, ps.ByName("device")})
	}
}
