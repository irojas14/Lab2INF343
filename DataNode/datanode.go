package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	funcs "github.com/irojas14/Lab2INF343/Funciones"
	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	dn1Port = ":50055"
	dn2Port = ":50056"
	dn3Port = ":50057"
	local = "localhost"
	dn1Addrs = "dist149.inf.santiago.usm.cl" + dn1Port
	dn2Addrs = "dist151.inf.santiago.usm.cl" + dn1Port
	dn3Addrs = "dist152.inf.santiago.usm.cl" + dn1Port
)


type server struct {
	pb.UnimplementedDataNodeServer
}

func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
	fmt.Println("Solicitud de Registro de Jugadas")


	// Crear nuevo archivo de jugador, si es nuevo

	return &pb.RespuestaRegistrarJugadas{}, nil
}

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverJugadas) (*pb.RespuestaDevolverJugadas, error) {
	fmt.Println("Solicitud de Devolución de Jugadas")
	return &pb.RespuestaDevolverJugadas{}, nil
}


//archivo : jugador_numero__ronda_numero.txt

func CrearArchivoDeJugadas(numjugador, rondajugador int32){
	var NombreArchivo string = "jugador_" + funcs.FormatInt32(numjugador) + "__ronda_" + funcs.FormatInt32(rondajugador) + ".txt"
	funcs.CrearArchivoTxt(NombreArchivo)
	//funcs.InsertarJugadasDelJugador()
}

func main(){
	var srvAddr string;
	argsLen := len(os.Args)
	if argsLen > 3 {
		fmt.Println("Demasiados argumentos")
		return;
	} else if argsLen == 3 {
		srvAddr = local

		if (os.Args[1] == "1") {
			srvAddr += dn1Port
		} else if (os.Args[1] == "2") {
			srvAddr += dn2Port
		} else if (os.Args[1] == "3") {
			srvAddr += dn3Port
		}

	} else if argsLen < 3 {

		if (os.Args[1] == "1") {
			srvAddr = dn1Addrs
		} else if (os.Args[1] == "2") {
			srvAddr = dn2Addrs
		} else if (os.Args[1] == "3") {
			srvAddr = dn3Addrs
		}
	}
	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	pb.RegisterDataNodeServer(s, &server{})
	log.Printf("DataNode escuchando en %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}