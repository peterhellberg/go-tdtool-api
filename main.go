package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/julienschmidt/httprouter"
)

var (
	devMode = flag.Bool("d", false, "Development mode (fake tdtool output)")

	devicePattern = regexp.MustCompile(`(\d+)\t(.+)\t(ON|OFF)`)
)

type list struct {
	Count   int
	Devices []device
}

type device struct {
	ID     string
	Name   string
	Status bool
}

const exampleListOutput = `Number of devices: 4
1	Skrivbordet	OFF
5	Sovrum	OFF
2	Hörnlampa	OFF
4	Alla lampor	OFF
`

func listOutput() ([]byte, error) {
	if *devMode {
		return []byte(exampleListOutput), nil
	}

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

func main() {
	flag.Parse()

	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if out, err := listOutput(); err == nil {
			list := newList(out)

			index.Execute(w, list)
		}
	})

	router.GET("/list.json", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

var index = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Go TDTool API</title>
		<style type="text/css">
			.block-group,.block,.block-group:after,.block:after,.block-group:before,.block:before{-webkit-box-sizing:border-box;-moz-box-sizing:border-box;box-sizing:border-box}
			.block-group{*zoom:1}
			.block-group:before,.block-group:after{display:table;content:"";line-height:0}
			.block-group:after{clear:both}
			.block-group{list-style-type:none;padding:0;margin:0}
			.block-group>.block-group{clear:none;float:left;margin:0 !important}
			.block{float:left;width:100%}

			.col1 { width: 50%; }
			.col2 { width: 50%; }

			h1, h2 { font-family: Helvetica; margin-bottom: 0.1em; }
			h1 { font-size: 20px; }
			h2 { font-size: 15px; }

			button {
				display: block;
				width: 100%;
				height: 4em;
				border: 5px solid #fff;
				margin-bottom: 0.5em;
				font-size: 1em;
				background: #eee;
			}
		</style>
	</head>
	<body>
	<h1>Go TDTool API</h1>
	{{range .Devices}}
	<div class="block-group">
		<h2>{{.Name}}</h2>
		<div class="block col1">
			<div class="box">
				<button onclick="put('/{{.ID}}/on')">ON</button>
			</div>
		</div>
		<div class="block col2">
			<div class="box">
				<button onclick="put('/{{.ID}}/off')">OFF</button>
			</div>
		</div>
	</div>
	{{end}}
	<script>
		function put(url){
			xmlhttp=new XMLHttpRequest();
			xmlhttp.open("PUT",url,true)
			xmlhttp.send(null);
		}
	</script>
	</body>
</html>`))
