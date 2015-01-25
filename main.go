package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/julienschmidt/httprouter"
)

var devicePattern = regexp.MustCompile(`(\d+)\t(.+)\t(ON|OFF)`)

func listOutput() ([]byte, error) {
	return exec.Command("tdtool", "-l").Output()
}

func newList(out []byte) *list {
	devices := []device{}

	if found := devicePattern.FindAllStringSubmatch(string(out), -1); found != nil {
		for _, f := range found {
			devices = append(devices, device{
				ID:     f[1],
				Name:   f[2],
				Status: f[3] == "ON",
			})
		}
	}

	return &list{Count: len(devices), Devices: devices}
}

type list struct {
	Count   int
	Devices []device
}

type device struct {
	ID     string
	Name   string
	Status bool
}

func main() {
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if out, err := listOutput(); err == nil {
			list := newList(out)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(list)
		}
	})

	router.GET("/list", handle(Output, []string{"tdtool", "-l"}))

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
