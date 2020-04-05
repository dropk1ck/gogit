package main

import (
    //"flag"
    "fmt"
    "log"
    "os"
)

func CreateDir(name string) {
    err := os.Mkdir(name, 0755)
    if err != nil {
        log.Fatalf("ERROR: cannot create folder %s - %s", name, err)
    }
}

func CreateDirAll(path string) {
    err := os.MkdirAll(path, 0755)
    if err != nil {
        log.Fatalf("ERROR: cannot create folder %s - %s", path, err)
    }
}

func TouchFile(filepath string) {
    file, err := os.Create(filepath)
    if err != nil {
        log.Fatalf("ERROR: cannot create file %s - %s", filepath, err)
    }
    file.Close()
}

func InitRepo() {
    // the git init command creates a directory file hierarchy that looks like this:
    //   .git/
    //   .git/objects/
    //   .git/refs/
    //   .git/refs/heads/
    //   .git/refs/tags/
    //   .git/HEAD
    //   .git/config
    //   .git/description

    // the following func calls are sufficent to create an empty repo, all dirs in the
    // path will automatically be created if they don't exist
    CreateDirAll(".git/objects/")
    CreateDirAll(".git/refs/heads/")
    CreateDirAll(".git/refs/tags/")
    TouchFile(".git/HEAD")
    TouchFile(".git/config")
    TouchFile(".git/description")
    fmt.Println("Initialized empty repo at .git/")
}

func main() {
    //initCmd := flag.NewFlagSet("init", flag.ExitOnError)

    if len(os.Args) < 2 {
        fmt.Println("no subcommand specified")
        os.Exit(1)
    }

    subcommand := os.Args[1]
    switch subcommand {
    case "init":
        InitRepo()
    default:
        fmt.Println("unknown subcommand: %s", subcommand)
    }
}
