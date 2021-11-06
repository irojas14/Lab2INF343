package main

import (
	"bufio"
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
	fmt.Printf("Almacenar Jugada %v del Jugador %v del Juego %v\n", 
	in.JugadasJugador.JugadasJuego[0].Jugadas[0], in.JugadasJugador.NumJugador.Val, in.JugadasJugador.JugadasJuego[0].NumJuego)

	jugadorStr := strconv.FormatInt(int64(in.JugadasJugador.NumJugador.Val), 10)
	etapaStr := strconv.FormatInt(int64(in.JugadasJugador.JugadasJuego[0].NumJuego), 10)

	var nombreArchivo = "jugador_" + jugadorStr + "__Etapa_" + etapaStr + ".txt"

	// Crear nuevo archivo de jugador, si es nuevo
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

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverJugadasDataNode) (*pb.RespuestaDevolverJugadas, error) {
	fmt.Println("Solicitud de Devolución de Jugadas")
	
	jugadorStr := strconv.FormatInt(int64(in.NumJugador.Val), 10)

	prefix := "jugador_" + jugadorStr + "__Etapa_"

	res := &pb.RespuestaDevolverJugadas{
		JugadasJugador: &pb.JugadasJugador{
			NumJugador: in.NumJugador,
		},
	}

	for _, etapaStr := range in.Etapas {
		nombreArchivo := prefix + etapaStr + ".txt"
		jugadas, err := ObtenerJugadasArray(nombreArchivo)
		if (err != nil) {
			log.Fatalf("Error al buscar las jugadas en el archivo: err: %v\n", err)
			return nil, err
		}

		numJuego32, err2 := strconv.ParseInt(etapaStr, 10, 32)

		if (err2 != nil) {
			log.Fatalf("Error en la conversión al buscar las jugadas en el archivo: err: %v\n", err2)			
			return nil, err2
		}

		jugadasJuego := &pb.JugadasJuego {
			NumJuego: pb.JUEGO(numJuego32),
			Jugadas: jugadas,
		}
		res.JugadasJugador.JugadasJuego = append(res.JugadasJugador.JugadasJuego, jugadasJuego)
	}
	return res, nil
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

func ObtenerJugadasArray(nombreArchivo string) ([]*pb.Jugada, error) {
	file, err := os.Open("DataNode/" + curFiles + "/" + nombreArchivo)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	var jugadas []*pb.Jugada
	for scanner.Scan() {
		jStr := scanner.Text()
		j, err2 := strconv.ParseInt(jStr, 10, 32)
		if (err2 != nil)  {
			log.Fatalf("Error en la Conversion de la Jugada: err: %v\n", err2)
			return nil, err2
		}
		j32 := int32(j)
		jugada := &pb.Jugada{Val: j32}
		jugadas = append(jugadas, jugada)
	}
	return jugadas, nil
}