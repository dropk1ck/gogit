package main

import (
    "flag"
    "bytes"
    "compress/zlib"
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "gopkg.in/ini.v1"
)

// the basic object in git: a type, contents, and size of the contents.
// really just a TLV with extra bytes between them (when the object is serialized)
type Object struct {
    Type string
    Size uint64
    Contents string
}

func NewObject(objtype string, contents string) *Object {
    obj := new(Object)
    obj.Type = objtype
    obj.Size = uint64(len(contents))
    obj.Contents = contents
    return obj
}

// serializing a git object:
//   - serialized object format:
//     - header (blob, commit, tag, tree)
//     - ASCII space (0x20)
//     - size of object in bytes as ASCII number
//     - NULL (0x00)
//     - object contents

func (obj Object) Serialize() []byte {
    var b bytes.Buffer

    // convert our object to bytes
    b.WriteString(obj.Type)
    b.WriteString(" ")
    b.WriteString(strconv.FormatUint(obj.Size, 10))
    b.WriteString(string(0x00))
    b.WriteString(obj.Contents)
    return b.Bytes()
}

// stores a serialized git object to the repository
//   - take sha1 sum of contents
//   - first 2 bytes are dirname
//   - remaining bytes are filename in dir
//   - objects are stored *compressed* with zlib

func WriteObjectToFile(objBytes []byte, sha1Sum []byte) {
    // compress the bytes into a buffer
    var zb bytes.Buffer
    w := zlib.NewWriter(&zb)
    w.Write(objBytes)
    w.Close()

    // create the directory paths as needed
    sha1Hash := hex.EncodeToString(sha1Sum[:])
    CreateDirAll(".git/objects/" + sha1Hash[0:2])

    err := ioutil.WriteFile(".git/objects/" + sha1Hash[0:2] + "/" + sha1Hash[2:], zb.Bytes(), 0644)
    if err != nil {
        log.Fatalf("ERROR: could not write object %s - %s", sha1Hash[2:], err)
    }
}

func ReadObjectFromFile(sha1Hash string) {
    //var contents bytes.Buffer

    filePath := ".git/objects/" + sha1Hash[0:2] + "/" + sha1Hash[2:]
    zContents, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Fatalf("ERROR: cannot read file %s - %s", filePath, err)
    }

    // contents are compress with zlib, uncompress
    zReader := bytes.NewReader(zContents)
    r, err := zlib.NewReader(zReader)
    if err != nil {
        log.Fatalf("ERROR: could not decompress contents - %s", err)
    }
    io.Copy(os.Stdout, r)
    r.Close()
}

func HashObject(path string, storeObject bool, tag string) {
    // suck file contents into memory
    contents, err := ioutil.ReadFile(path)
    if err != nil {
        log.Fatalf("ERROR: cannot read file %s - %s", path, err)
    }

    // create an object and serialize it
    obj := NewObject(tag, string(contents))
    objBytes := obj.Serialize()

    // compute the SHA1 hash
    h := sha1.New()
    h.Write(objBytes)
    sha1sum := h.Sum(nil)

    fmt.Printf("object SHA1 hash: %x\n", sha1sum)

    // optionally write object to repo, compressing with zlib
    if (storeObject) {
        WriteObjectToFile(objBytes, sha1sum)
    }
}

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
    //   - bare indicates whether or not a worktree is present from what I understand
    config, _ := ini.Load(".git/config")
    config.Section("core").Key("repositoryformatversion").SetValue("0")
    config.Section("core").Key("filemode").SetValue("false")
    config.Section("core").Key("bare").SetValue("false")
    config.SaveTo(".git/config")

    fmt.Println("Initialized empty repo at .git/")
}

func main() {
    //initCmd := flag.NewFlagSet("init", flag.ExitOnError)
    hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
    hashObjStore := hashObjCmd.Bool("w", false, "write object to repository")
    hashObjTag := hashObjCmd.String("t", "blob", "type of object tag")

    catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

    if len(os.Args) < 2 {
        fmt.Println("no subcommand specified")
        os.Exit(1)
    }

    subcommand := os.Args[1]
    switch subcommand {
    case "init":
        InitRepo()
    case "hash-object":
        hashObjCmd.Parse(os.Args[2:])
        args := hashObjCmd.Args()
        if len(args) < 1 {
            fmt.Println("hash-object: expected path to file")
            os.Exit(1)
        }
        path := args[0]
        HashObject(path, *hashObjStore, *hashObjTag)
    case "cat-file":
        catFileCmd.Parse(os.Args[2:])
        args := catFileCmd.Args()
        if len(args) < 2 {
            fmt.Println("cat-file: expected type and object SHA1 hash")
            os.Exit(1)
        }
        sha1Hash := args[1]
        ReadObjectFromFile(sha1Hash)
    default:
        fmt.Println("unknown subcommand: %s", subcommand)
    }
}
