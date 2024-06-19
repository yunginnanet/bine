package main

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/cretz/bine/process/embedded"
)

func main() {
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
	if err := embedded.Init(torStaticPath); err != nil {
		panic(err)
	}
	oldF, err := os.OpenFile("process/embedded/process.go", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	newPath := filepath.Join(os.TempDir(), "process.go")
	newF, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, 0644)
	xerox := bufio.NewScanner(oldF)
	doNext := false
	for xerox.Scan() {
		if xerox.Err() != nil {
			panic(xerox.Err())
		}
		if strings.Contains(xerox.Text(), "NewCreator() process.Creator") {
			doNext = true
		}
		if doNext {
			if _, err = newF.WriteString("	return tor047.NewCreator()\n"); err != nil {
				panic(err)
			}
			_ = xerox.Scan() // ignore the next line
			doNext = false
			continue
		}
		if _, err = newF.WriteString(xerox.Text() + "\n"); err != nil {
			panic(err)
		}
	}
	if err = newF.Sync(); err != nil {
		panic(err)
	}
	if err = newF.Close(); err != nil {
		panic(err)
	}
	if err = oldF.Close(); err != nil {
		panic(err)
	}
	if err = os.Rename(newPath, oldF.Name()); err != nil {
		panic(err)
	}
}
