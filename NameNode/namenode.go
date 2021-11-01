package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"
	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port    = ":50054"
	local   = "localhost" + port
	address = "dist149.inf.santiago.usm.cl" + port
)

const (
	dnPort = ":50055"
	dn1Addrs = "dist150.inf.santiago.usm.cl" + dnPort
	dn2Addrs = "dist151.inf.santiago.usm.cl" + dnPort
	dn3Addrs = "dist152.inf.santiago.usm.cl" + dnPort
)

const (
	dnLocal = "localhost"
	local1 = dnLocal + ":50055"
	local2 = dnLocal + ":50056"
	local3 = dnLocal + ":50057"
)

var (
	RemoteAddrs [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs} 
	localAddrs [3]string = [3]string{local1, local2, local3}
	curAddrs [3]string
)

var DataNodeAddresses [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs}

var UbicacionJugadasJugadores = "UbicaciónJugadasDeJugadores.txt"


type server struct {
	pb.UnimplementedNameNodeServer
}


func LeerRegistroDeJugadas(numjugador int32) (*pb.JugadasJugador, error) {

	file, ferr := os.Open(UbicacionJugadasJugadores)
	if ferr != nil {
		panic(ferr)
	}

	scanner := bufio.NewScanner(file)

	res := &pb.JugadasJugador{}
	res.NumJugador = &pb.JugadorId{Val: numjugador}

	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, " ")
		//items[0] = Jugador_numero
		//items[1] = Ronda_numero
		//items[2] = ip_datanode
		if items[0] == "Jugador_" + funcs.FormatInt32(numjugador) {
			fmt.Println(items)

			conn, err := grpc.Dial(items[2], grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
				return nil, err;
			}
			defer conn.Close()
			dc := pb.NewDataNodeClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r, err := dc.DevolverJugadas(ctx, &pb.SolicitudDevolverJugadas{NumJugador: &pb.JugadorId{Val: numjugador}})
			if err != nil {
				log.Fatalf("Error: %v\n", err)
				return nil, err;
			}
			//r.JugadasJugador.NumJugador.GetVal()
			//r.JugadasJugador.JugadasRonda[0].NumRonda.GetVal()
			//r.JugadasJugador.JugadasRonda[0].Jugadas[0].GetVal()
			res.JugadasRonda = append(res.JugadasRonda, r.JugadasJugador.JugadasRonda[0])
		}
	}
	return res, nil
}


func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
    log.Println("Sirviendo Solicitud de Registrar Jugada")

	// Elegir un Datanode y guarda la info
	// creado un DataNodeClient y usando la RPC DataNode.RegistrarJugadas.

    return &pb.RespuestaRegistrarJugadas{ NumJugador: &pb.JugadorId{Val: in.JugadasJugador.NumJugador.Val} }, nil
}

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverJugadas) (*pb.RespuestaDevolverJugadas, error) {
	log.Println("Sirviendo Solicitud de Devolver Jugada")
	var numJugador = in.GetNumJugador().Val

	// jj = jugadas jugador
	jj, err := LeerRegistroDeJugadas(numJugador);

	if err == nil {
		return nil, err;
	}
	return &pb.RespuestaDevolverJugadas{JugadasJugador: jj}, nil
}

func main() {	
	curAddrs = RemoteAddrs
	fmt.Println("Comenzando El NameNode")	
	if len(os.Args) == 2 {
		curAddrs = localAddrs
	}

	for _, elem := range(curAddrs) {
		fmt.Printf(elem + " - ")
	}
	fmt.Println()

	// Set up a connection to the server.
	conn, err := grpc.Dial(curAddrs[0], grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewDataNodeClient(conn)

	r, err := c.RegistrarJugadas(context.Background(), &pb.SolicitudRegistrarJugadas{})

	if err != nil {
		log.Fatalf("Error en la respuesta: %v\n", err)
	}
	fmt.Println("Respuesta Recibida: " + r.String())

	// Set up a connection to the server.
	conn, err = grpc.Dial(curAddrs[1], grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	c = pb.NewDataNodeClient(conn)

	r, err = c.RegistrarJugadas(context.Background(), &pb.SolicitudRegistrarJugadas{})

	if err != nil {
		log.Fatalf("Error en la respuesta: %v\n", err)
	}
	fmt.Println("Respuesta Recibida: " + r.String())
}

