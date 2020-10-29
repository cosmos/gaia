package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/cosmos/gaia/cmd/gaiad/cmd"
)

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	go profile()

	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}

func profile() {
	timer := time.NewTimer(4 * time.Second)
	<-timer.C
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Can't get current directory.", err)
	}
	fn := filepath.Join(cwd, "cpu.pprof")

	log.Println("Starting collecting CPU profile to ", fn)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("profile: could not create cpu profile %q: %v", fn, err)
	}
	pprof.StartCPUProfile(f)

	timer = time.NewTimer(35 * time.Minute)
	<-timer.C
	pprof.StopCPUProfile()
	f.Close()
	log.Println("CPU profile stopped")
}
