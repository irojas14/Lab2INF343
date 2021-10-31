package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port    = ":50052"
	local   = "localhost" + port
	address = "dist150.inf.santiago.usm.cl" + port
)


var JugadoreEliminados = "JugadoresEliminados.txt"
var MontoAcumulado int32 = 0

type server struct {
	pb.UnimplementedPozoServer
}

func (s *server) VerMonto(ctx context.Context, in *pb.SolicitudVerMonto) (*pb.RespuestaVerMonto, error) {
    return &pb.RespuestaVerMonto{ Monto: float32(MontoAcumulado) }, nil
}

func main() {
	srvAddr := address
	if len(os.Args) == 2 {
		srvAddr = local
	}

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("faled to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterPozoServer(s, &server{})
	log.Printf("Pozo escuchando en %v", lis.Addr())
	
    if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

    /*
    createFile()
    JugadorEliminado()
    readFile()
    deleteFile()
    */
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

func JugadorEliminado() {
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