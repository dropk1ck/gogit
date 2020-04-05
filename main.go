package main

import (
    //"flag"
    "fmt"
    "log"
    "os"
    "gopkg.in/ini.v1"
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

    // the following code is sufficent to create an empty repo, all dirs in the
    // path will automatically be created if they don't exist
    CreateDirAll(".git/objects/")
    CreateDirAll(".git/refs/heads/")
    CreateDirAll(".git/refs/tags/")

    TouchFile(".git/HEAD")
    TouchFile(".git/config")
    TouchFile(".git/description")

    // bare-bones ini config
    //   - repositoryformatversion is the version of the gitdir format, 0 is typical, 1 is with extensions?
    //   - filemode controls tracking the file mode changes in the tree
    //   - bare... I'll look this up later, something to do with worktrees
    config, _ := ini.Load(".git/config")
    config.Section("core").Key("repositoryformatversion").SetValue("0")
    config.Section("core").Key("filemode").SetValue("false")
    config.Section("core").Key("bare").SetValue("false")
    config.SaveTo(".git/config")

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
