package main

import (
    "flag"
    "bytes"
    "compress/zlib"
    "crypto/sha1"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "gopkg.in/ini.v1"
)

// the basic object in git. a type, contents, and size of the contents.
// really just a TLV with extra bytes between them
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

// git object:
//   - take sha1 sum of contents
//   - first 2 bytes are dirname
//   - remaining bytes are filename in dir
//   - object format:
//     - header (blob, commit, tag, tree)
//     - ASCII space (0x20)
//     - size of object in bytes as ASCII number
//     - NULL (0x00)
//     - object contents
//   - objects are stored *compressed* with zlib

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

func WriteObjectToFile(objBytes []byte, sha1sum []byte) {
    // compress the bytes into a buffer
    var zb bytes.Buffer
    w := zlib.NewWriter(&zb)
    w.Write(objBytes)
    w.Close()

    // create the directory paths as needed
}

func HashObject(path string, storeObject bool, tag string) {
    // suck file contents into memory
    contents, err := ioutil.ReadFile(path)
    if err != nil {
        log.Fatalf("ERROR: cannot read file %s - %s", path, err)
    }

    // serialize object
    obj := NewObject(tag, string(contents))
    objBytes := SerializeObject(obj)

    // compute the SHA1 hash
    h := sha1.New()
    h.Write(objBytes)
    sha1sum := h.Sum(nil)

    fmt.Printf("Serialized object SHA1 hash: %x\n", sha1sum)

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
    default:
        fmt.Println("unknown subcommand: %s", subcommand)
    }
}
