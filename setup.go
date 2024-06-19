package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

var skipCompile = false

func init() {
	flag.BoolVar(&skipCompile, "skip-compile", false, "Skip compiling tor-static")
	flag.Parse()
	torStaticPath := os.Getenv("TOR_STATIC_PATH")
	if torStaticPath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		surroundings, err := os.ReadDir(pwd)
		if err != nil {
			panic(err)
		}
		for _, file := range surroundings {
			if file.IsDir() && file.Name() == "tor-static" {
				torStaticPath = file.Name()
				break
			}
			if file.IsDir() && file.Name() == "process" {
				torStaticPath = filepath.Join(pwd, "process", "embedded", "tor-static")
				break
			}
			if file.IsDir() && file.Name() == "embedded" {
				torStaticPath = filepath.Join(pwd, "embedded", "tor-static")
				break
			}
			if file.IsDir() && file.Name() == "tor-0.4.7" {
				torStaticPath = filepath.Join(pwd, "tor-static")
			}
		}
		if skipCompile {
			goto fixProcess
		}
	clone:
		if _, err = git.PlainClone(torStaticPath, false, &git.CloneOptions{
			URL:               "https://github.com/yunginnanet/tor-static",
			Progress:          os.Stdout,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			println("Failed to clone tor-static")
			println(err.Error())
			os.Exit(1)
		} else if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			presentFiles, err := os.ReadDir(torStaticPath)
			if err != nil {
				println("Failed to read tor-static directory")
				println(err.Error())
				os.Exit(1)
			}
			if len(presentFiles) < 2 {
				if err = os.RemoveAll(torStaticPath); err != nil {
					println("Failed to remove tor-static directory")
					println(err.Error())
					os.Exit(1)
				}
				goto clone
			}
		}
		cmd := exec.Command("go", "run", "build.go", "-verbose", "build-all")
		cmd.Dir = torStaticPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			println("Failed to build tor-static")
			println(err.Error())
			os.Exit(1)
		}
	}
	/*	if err := embedded.Init(torStaticPath); err != nil {
		panic(err)
	}*/
fixProcess:
	oldFPath := "./process/embedded/process.go"
	println("Fixing process.go at " + oldFPath)
	oldData, err := os.ReadFile(oldFPath)
	if err != nil {
		panic(err)
	}
	oldF := bytes.NewReader(oldData)
	newPath := filepath.Join(os.TempDir(), "bine"+strconv.Itoa(int(time.Now().Unix())), "process.go")
	if err = os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		panic(err)
	}
	newF, err := os.Create(newPath)
	if err != nil {
		panic(err)
	}
	xerox := bufio.NewScanner(oldF)
	doNext := false
	inImports := false
	for xerox.Scan() {
		if xerox.Err() != nil {
			panic(xerox.Err())
		}
		if strings.Contains(xerox.Text(), "import (") {
			// println("Found imports")
			inImports = true
		}
		if inImports && strings.Contains(xerox.Text(), ")") {
			inImports = false
			toWrite := `"github.com/yunginnanet/bine/process/embedded/tor-0.4.7"
			)`
			// println("Writing(+): " + toWrite)
			if _, err = newF.WriteString(toWrite); err != nil {
				panic(err)
			}
			continue
		}

		needle := "NewCreator() process.Creator"
		if strings.Contains(xerox.Text(), needle) {
			doNext = true
			// println("found NewCreator")
			_, _ = newF.WriteString(xerox.Text() + "\n")
			continue
		}
		if doNext {
			toWrite := strings.ReplaceAll(
				xerox.Text(),
				"return nil",
				"return tor047.NewCreator()",
			)
			// println("Writing(+): " + toWrite)
			if _, err = newF.WriteString(toWrite + "\n"); err != nil {
				panic(err)
			}
			doNext = false
			continue
		}
		toWrite := xerox.Text() + "\n"
		// println("Writing: " + toWrite)
		if _, err = newF.WriteString(toWrite); err != nil {
			panic(err)
		}
	}
	if err = newF.Sync(); err != nil {
		panic(err)
	}
	if err = newF.Close(); err != nil {
		panic(err)
	}
	if err = exec.Command("gofmt", "-w", newPath).Run(); err != nil {
		dat, e := os.ReadFile(newPath)
		if e == nil {
			println(fmt.Sprintf("Failed to format process.go, dumping %s...\n\n", newPath))
			_, _ = os.Stdout.Write(dat)
			println("\n\n")
		}
		panic(err)
	}
	var oldFPathAbs string
	if oldFPathAbs, err = filepath.Abs(oldFPath); err != nil {
		panic(err)
	}
	if _, err = os.Stat(filepath.Join(filepath.Dir(oldFPathAbs), "process.go.bak")); errors.Is(err, os.ErrNotExist) {
		if err = os.Rename(oldFPathAbs, filepath.Join(filepath.Dir(oldFPathAbs), "process.go.bak")); err != nil {
			panic(err)
		}
	}
	newDat, err := os.ReadFile(newPath)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(oldFPathAbs, newDat, 0644); err != nil {
		_ = os.Rename(filepath.Join(filepath.Dir(oldFPathAbs), "process.go.old"), oldFPathAbs)
		panic(err)
	}
	_ = os.RemoveAll(filepath.Dir(newPath))
}
