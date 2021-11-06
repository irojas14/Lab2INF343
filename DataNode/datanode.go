package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	funcs "github.com/irojas14/Lab2INF343/Funciones"
	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	dn1Port = ":50057"
	dn2Port = ":50058"
	dn3Port = ":50059"
	local = "localhost"
	dn1Addrs = "dist149.inf.santiago.usm.cl" + dn1Port
	dn2Addrs = "dist151.inf.santiago.usm.cl" + dn2Port
	dn3Addrs = "dist152.inf.santiago.usm.cl" + dn3Port
)

var (
	filesDn1 = "FilesDataNode1"
	filesDn2 = "FilesDataNode2"
	filesDn3 = "FilesDataNode3"
	curFiles = "DataNode/"
)


type server struct {
	pb.UnimplementedDataNodeServer
}

func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
	fmt.Println("Solicitud de Registro de Jugadas")
	// Crear nuevo archivo de jugador, si es nuevo

	fmt.Printf("Almacenar Jugada del Jugador %v del Juego %v\n", in.JugadasJugador.NumJugador.Val, in.JugadasJugador.JugadasJuego[0].NumJuego)

	jugadorStr := strconv.FormatInt(int64(in.JugadasJugador.NumJugador.Val), 10)
	etapaStr := strconv.FormatInt(int64(in.JugadasJugador.JugadasJuego[0].NumJuego), 10)

	var nombreArchivo = "jugador_" + jugadorStr + "__Etapa_" + etapaStr + ".txt"
	if (!funcs.Is_in_folder(nombreArchivo, curFiles)) {
		funcs.CrearArchivoTxt(curFiles + "/" + nombreArchivo)
	}

	// Obtener la jugada

	jugada := in.JugadasJugador.JugadasJuego[0].Jugadas[0].Val
	jugadas := []int32{jugada}
	
	// insertarla en el archivo correspondiente
	funcs.InsertarJugadasDelJugador(curFiles + "/" + nombreArchivo, jugadas)

	fmt.Printf("Se Insertó la Jugada: %v del Jugador %v en el Juego: %v en el archivo %v\n", 
	jugadas[0], in.JugadasJugador.NumJugador.Val, in.JugadasJugador.JugadasJuego[0].NumJuego, nombreArchivo)

	fmt.Println("Retornando")
	return &pb.RespuestaRegistrarJugadas{NumJugador: in.JugadasJugador.NumJugador}, nil
}

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverJugadas) (*pb.RespuestaDevolverJugadas, error) {
	fmt.Println("Solicitud de Devolución de Jugadas")
	return &pb.RespuestaDevolverJugadas{}, nil
}

// is_in_folder(file_name string, ruta string)
/*
	DataNode
		//FilesDataNode1
		//FilesDataNode2
		//FilesDataNode3
*/
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
			curFiles += filesDn1
		} else if (os.Args[1] == "2") {
			srvAddr += dn2Port
			curFiles += filesDn2
		} else if (os.Args[1] == "3") {
			srvAddr += dn3Port
			curFiles += filesDn3
		}

	} else if argsLen < 3 {

		if (os.Args[1] == "1") {
			srvAddr = dn1Addrs
			curFiles += filesDn1
		} else if (os.Args[1] == "2") {
			srvAddr = dn2Addrs
			curFiles += filesDn2			
		} else if (os.Args[1] == "3") {
			srvAddr = dn3Addrs
			curFiles += filesDn3			
		}
	}

	fmt.Printf("Carpeta CurFiles: %v\n",curFiles)
	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	pb.RegisterDataNodeServer(s, &server{})
	log.Printf("DataNode escuchando en %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}