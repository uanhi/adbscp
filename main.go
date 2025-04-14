package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type config struct {
	port       string
	device     string
	useDefault bool
}

var conf config

func init() {
	port := flag.String("port", ":8080", "")
	device := flag.String("device", "dbcfdab4", "")

	useDefault := flag.Bool("def", false, "")

	flag.Usage = func() {
		fmt.Println(`adbscp (Real-time ADB screenshot preview in browser)
	
Usage:
  adbscp --port :8080 --device <serial>
	
Options:
  --port <port>      port number to listen on (e.g. :8080)
  --device <serial>  adb device serial number`)
	}
	flag.Parse()

	conf.port = *port
	conf.device = *device
	conf.useDefault = *useDefault
}

func main() {
	if len(os.Args) == 1 && !conf.useDefault {
		flag.Usage()
		return
	}
	exists, err := deviceExists(conf.device)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if !exists {
		fmt.Printf("device not found %s\n", conf.device)
		return
	}
	http.HandleFunc("/", Get)

	if err := http.ListenAndServe(conf.port, nil); err != nil {
		panic(err)
	}
}

func deviceExists(serial string) (bool, error) {
	cmd := exec.Command("adb", "devices")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, err
	}
	lines := strings.SplitSeq(out.String(), "\n")

	for line := range lines {
		part := strings.Split(line, "\t")
		if part[0] == serial && part[1] == "device" {
			return true, nil
		}
	}
	return false, nil
}

func Get(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}
	err := exec.Command(
		"adb",
		"-s", conf.device,
		"shell", "screencap", "-p", "/sdcard/screencap.png",
	).Run()

	if err != nil {
		log.Println(err)
		http.Error(w, "screencap failed", http.StatusInternalServerError)
		return
	}
	err = exec.Command(
		"adb",
		"-s", conf.device,
		"pull", "/sdcard/screencap.png", ".",
	).Run()

	if err != nil {
		log.Println(err)
		http.Error(w, "pull failed", http.StatusInternalServerError)
		return
	}
	log.Println("OK")
	http.ServeFile(w, r, "./screencap.png")
}
