package main

import (
	"fmt"
	"io"
	"os"
)

const (
	port    = ":50052"
	local   = "localhost" + port
	address = "dist150.inf.santiago.usm.cl" + port
)


var JugadoreEliminados = "JugadoresEliminados.txt"

func main() {
    createFile()
    writeFile()
    readFile()
    deleteFile()
}

func createFile() {
    // check if file exists
    var _, err = os.Stat(JugadoreEliminados)

    // create file if not exists
    if os.IsNotExist(err) {
        var file, err = os.Create(JugadoreEliminados)
        if isError(err) {
            return
        }
        defer file.Close()
    }

    fmt.Println("File Created Successfully", JugadoreEliminados)
}

func writeFile() {
    // Open file using READ & WRITE permission.
    var file, err = os.OpenFile(JugadoreEliminados, os.O_RDWR, 0644)
    if isError(err) {
        return
    }
    defer file.Close()

    // Write some text line-by-line to file.
    _, err = file.WriteString("Hello \n")
    if isError(err) {
        return
    }
    _, err = file.WriteString("World \n")
    if isError(err) {
        return
    }

    // Save file changes.
    err = file.Sync()
    if isError(err) {
        return
    }

    fmt.Println("File Updated Successfully.")
}

func readFile() {
    // Open file for reading.
    var file, err = os.OpenFile(JugadoreEliminados, os.O_RDWR, 0644)
    if isError(err) {
        return
    }
    defer file.Close()

    // Read file, line by line
    var text = make([]byte, 1024)
    for {
        _, err = file.Read(text)

        // Break if finally arrived at end of file
        if err == io.EOF {
            break
        }

        // Break if error occured
        if err != nil && err != io.EOF {
            isError(err)
            break
        }
    }

    fmt.Println("Reading from file.")
    fmt.Println(string(text))
}

func deleteFile() {
    // delete file
    var err = os.Remove(JugadoreEliminados)
    if isError(err) {
        return
    }

    fmt.Println("File Deleted")
}

func isError(err error) bool {
    if err != nil {
        fmt.Println(err.Error())
    }

    return (err != nil)
}