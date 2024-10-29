package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	esbuild "github.com/evanw/esbuild/pkg/api"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

const Ver = "0.1"

var entrypoint string

func copyFile(dst string, src string) error {
	f1, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f1.Close()

	f2, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f2.Close()

	_, err = io.Copy(f2, f1)
	return err
}

type Util struct {
	Path string
}

func (u Util) Exec(args ...string) (s string, err error) {
	builder := new(strings.Builder)
	cmd := exec.Command(u.Path, args...)
	cmd.Stdout = builder
	err = cmd.Run()
	if err != nil {
		return
	}

	s = builder.String()
	return
}

func (u Util) Beacon(args ...string) error {
	return exec.Command(u.Path, args...).Run()
}

func (u Util) BeaconToStdout(args ...string) error {
	cmd := exec.Command(u.Path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (u Util) BeaconToStdoutInDir(dir string, args ...string) error {
	cmd := exec.Command(u.Path, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func esbuildPreset(wdir string, first bool, options esbuild.BuildOptions) {
	if first {
		options.EntryPoints = []string{entrypoint}
	} else {
		options.EntryPoints = []string{wdir + "/script"}
		options.AllowOverwrite = true
	}

	result := esbuild.Build(options)
	if len(result.Errors) != 0 {
		log.Printf("[esbuild] got %d error(s)!", len(result.Errors))
		for _, err := range result.Errors {
			log.Printf("[esbuild] %s", err.Text)
		}

		log.Fatal("[esbuild] not continuing!")
	}

	if len(result.OutputFiles) == 0 {
		log.Fatal("[esbuild] no files outputted")
	}

	err := os.WriteFile(wdir+"/script", result.OutputFiles[0].Contents, 0644)
	if err != nil {
		log.Fatal("[esbuild] failed to save script", "err", err)
	}
}

func terserPreset(wdir string, first bool, options []string) {
	terser := Util{"terser"}

	if first {
		options = append([]string{entrypoint}, options...)
	} else {
		options = append([]string{"script"}, options...)
	}
	options = append(options, "-o", "script")
	err := terser.BeaconToStdoutInDir(wdir, options...)
	if err != nil {
		log.Fatalf("[terser] got error %s", err)
	}
}

func qjscPreset(wdir string, first bool) {
	const moduleName = "f"
	qjsc := Util{"qjsc"}

	var sc string
	if first {
		sc = entrypoint
	} else {
		sc = wdir + "/script"
	}

	// copy the file so it has our desired module name (and doesn't leak the full script path on your machine)
	err := copyFile(wdir+"/"+moduleName, sc)
	if err != nil {
		log.Fatalf("[qjsc] failed to copy script: %s", err)
	}

	err = qjsc.BeaconToStdoutInDir(wdir, "-c", "-m", moduleName, "-o", "script")
	if err != nil {
		log.Fatalf("[qjsc] got error %s", err)
	}

	data, err := os.ReadFile(wdir + "/script")
	if err != nil {
		log.Fatalf("[qjsc] failed to read compiled script: %s", err)
	}

	data = bytes.ReplaceAll(bytes.ReplaceAll(bytes.Split(bytes.Split(data, []byte("{"))[1], []byte("}"))[0], []byte(" "), []byte{}), []byte("\n"), []byte{})
	b := bytes.Split(data, []byte(","))
	b = b[:len(b)-1]
	data = []byte{}
	for _, byt := range b {
		n, err := strconv.ParseUint(string(byt[2:]), 16, 8)
		if err != nil {
			log.Fatalf("[qjsc] failed to parse compiled script: %s", err)
		}
		data = append(data, byte(n))
	}

	err = os.WriteFile(wdir+"/script", data, 0644)
	if err != nil {
		log.Fatalf("[qjsc] failed to write compiled script: %s", err)
	}
}

type preset struct {
	ReuseWorkflow string

	Util           string
	Options        *[]string
	EsbuildOptions *esbuild.BuildOptions
}

var workflows = make(map[string][]preset)

func loadWorkflows(filename string) {
	if filename == "" {
		return
	}

	var f map[string][]preset
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("failed to read %s: %s", filename, err)
		return
	}

	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Printf("failed to unmarshal %s: %s", filename, err)
		return
	}

	for key, val := range f {
		workflows[key] = val
	}
}

func runWorkflow(workflow string, wdir string, nested *bool) string {
	flow, ok := workflows[workflow]
	if !ok {
		log.Fatalf("workflow %s not found!", workflow)
	}

	if len(flow) == 0 {
		log.Fatalf("workflow %s is empty!", workflow)
	}

	if wdir == "" {
		_wdir, err := os.MkdirTemp(cwd, ".frida_builder_")
		if err != nil {
			log.Fatalf("failed to make tempdir: %s", err)
		}

		wdir = _wdir
	}

	for i, set := range flow {
		var first bool
		if nested != nil {
			first = *nested
			nested = nil
		} else {
			first = i == 0
		}

		if set.ReuseWorkflow != "" {
			runWorkflow(set.ReuseWorkflow, wdir, &first)
			continue
		}

		switch set.Util {
		case "esbuild":
			if set.EsbuildOptions == nil {
				log.Fatalf("error in workflow %s: preset %d uses esbuild but provides no EsbuildOptions", workflow, i)
			}

			log.Print("running esbuild preset...")
			esbuildPreset(wdir, first, *set.EsbuildOptions)
			continue
		case "terser":
			var s []string
			if set.Options != nil {
				s = *set.Options
			}

			log.Print("running terser preset...")
			terserPreset(wdir, first, s)
			continue
		case "qjsc":
			log.Print("running qjsc preset...")
			qjscPreset(wdir, first)
			continue
		default:
			log.Fatalf("encountered unknown util %s (workflow %s, preset %d)", set.Util, workflow, i)
		}
	}

	return wdir
}

var bpath string
var cwd string

func main() {
	start := time.Now()
	flows := flag.String("workflows", "", "comma-separated list of paths to files with workflows")
	out := flag.String("o", "bundle", "where to output the bundle to")
	flag.Parse()
	log.Printf("running frida-builder v%s", Ver)

	{
		exec, err := os.Executable()
		if err != nil {
			log.Fatalf("failed to get own path: %s", err)
		}

		bpath = filepath.Dir(exec)

		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get working directory: %s", err)
		}

		cwd = wd
	}

	loadWorkflows(bpath + "/workflows.json")
	loadWorkflows("workflows.json")

	if flows != nil && *flows != "" {
		for _, flow := range strings.Split(*flows, ",") {
			loadWorkflows(flow)
		}
	}

	log.Printf("%d workflows loaded!", len(workflows))
	switch flag.Arg(0) {
	case "workflow":
		workflow := strings.ToLower(strings.Trim(flag.Arg(1), " "))
		entrypoint = flag.Arg(2)

		wdir := runWorkflow(workflow, "", nil)

		err := copyFile(*out, wdir+"/script")
		if err != nil {
			log.Fatalf("failed to copy output: %s", err)
		}

		os.RemoveAll(wdir)
		log.Printf("finished in %v!", time.Since(start))
	default:
		log.Fatal("unknown command")
	}
}
