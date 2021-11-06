package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"
	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port    = ":50054"
	local   = "localhost" + port
	address = "dist150.inf.santiago.usm.cl" + port
)

const (
	nameNodeFile = "registro.txt"
)

const (
	dnPort = ":50055"
	dn1Addrs = "dist149.inf.santiago.usm.cl" + dnPort
	dn2Addrs = "dist151.inf.santiago.usm.cl" + dnPort
	dn3Addrs = "dist152.inf.santiago.usm.cl" + dnPort
)

const (
	dnLocal = "localhost"
	local1 = dnLocal + ":50057"
	local2 = dnLocal + ":50058"
	local3 = dnLocal + ":50059"
)

var (
	RemoteAddrs [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs} 
	localAddrs [3]string = [3]string{local1, local2, local3}
	curAddrs [3]string
)

var DataNodeAddresses [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs}


type server struct {
	pb.UnimplementedNameNodeServer
}


func LeerRegistroDeJugadas(numjugador int32) (*pb.JugadasJugador, error) {

	file, ferr := os.Open("NameNode/" + nameNodeFile)
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
			res.JugadasJuego = append(res.JugadasJuego, r.JugadasJugador.JugadasJuego[0])
		}
	}
	return res, nil
}

func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
    log.Println("Sirviendo Solicitud de Registrar Jugada")

	// Elegir un Datanode y guarda la info
	// creado un DataNodeClient y usando la RPC DataNode.RegistrarJugadas

	// Elegimos un DataNode al Azar
	selDataNode := curAddrs[2];
	election := funcs.RandomInRange(1, 3)
	if (election == 1) {
		selDataNode = curAddrs[0]
	} else if (election == 2) {
		selDataNode = curAddrs[1]
	}

	// Escribirmos el Registro
	AgregarRegistro(in.JugadasJugador, selDataNode)

	// Almacenar Informacion al DataNode Seleccionado
	AlmacenamientoDataNode(in.JugadasJugador, selDataNode)

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

func AgregarRegistro(jugadasEnv *pb.JugadasJugador, selDataNodeAddrs string) error {
	jugador := jugadasEnv.NumJugador.Val
	juego := jugadasEnv.JugadasJuego[0].NumJuego
	
	jugadorStr := strconv.FormatInt(int64(jugador), 10)
	rondaStr := strconv.FormatInt(int64(juego), 10)

	var linea string = "Jugador " + jugadorStr + " Etapa " + rondaStr + " " + selDataNodeAddrs + "\n"

	file, err := os.Open("NameNode/" + nameNodeFile)

	if err != nil {
		fmt.Println(err)
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()+"\n"
		fmt.Printf("Line: %v - Linea: %v - iguales?: %v\n", line, linea, line == linea)
	
		if (linea == line) {
			fmt.Print("El archivo y el registro ya existe")
			file.Close()
			return nil
		}
	}
	file.Close()
	
	file2, err2 := os.OpenFile("NameNode/" + nameNodeFile, os.O_APPEND|os.O_WRONLY, 0777)

	if err != nil {
		fmt.Println(err)
		return err2
	}
	defer file2.Close()
	
	fmt.Printf("Linea registro: %v\n", linea)
	
	_, err3 := file2.WriteString(linea)

	if (err3 != nil ) {
		log.Fatalf("Error al escribir registro: %v\n", err3)
		return err3
	}
	return nil
}

func AlmacenamientoDataNode(jugadasEnv *pb.JugadasJugador, selDataNodeAddrs string) error {
	
	fmt.Printf("Conectándose a Datanode - Addr: %s\n", selDataNodeAddrs)
	
	// Set up a connection to the server.
	conn, err := grpc.Dial(selDataNodeAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
		return err
	}
	defer conn.Close()
	c := pb.NewDataNodeClient(conn)

	solicitudDeRegistro := &pb.SolicitudRegistrarJugadas{
		JugadasJugador: jugadasEnv,
	}
	r, err := c.RegistrarJugadas(context.Background(), solicitudDeRegistro)

	fmt.Println("Datanode Respondió")

	if (err != nil) {
		log.Fatalf("Error al registrar jugadas en el DataNode %v: err: %v\n", selDataNodeAddrs, err)
		return err
	}

	fmt.Printf("Se registró la jugada del Jugador: %v\n", r.NumJugador.Val)
	return nil
}

func main() {
	
	// Creamos el archivo de registro
	funcs.CrearArchivoTxt("NameNode/" + nameNodeFile)


	// Definimos nuestras direcciones: remotas o locales
	lisAddrs := address
	curAddrs = RemoteAddrs
	fmt.Println("Comenzando El NameNode")	
	if len(os.Args) == 2 {
		curAddrs = localAddrs
		lisAddrs = local
	}

	
	// Creamos el servidor que escuchara
	lis, err := net.Listen("tcp", lisAddrs)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterNameNodeServer(s, &server{})
	log.Printf("NameNode Inicializado: escuchando en %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}