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
	dn1Addrs = "dist149.inf.santiago.usm.cl" + ":50057"
	dn2Addrs = "dist151.inf.santiago.usm.cl" + ":50058"
	dn3Addrs = "dist152.inf.santiago.usm.cl" + ":50059"
)

const (
	dnLocal = "localhost"
	local1  = dnLocal + ":50057"
	local2  = dnLocal + ":50058"
	local3  = dnLocal + ":50059"
)

var (
	RemoteAddrs [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs}
	localAddrs  [3]string = [3]string{local1, local2, local3}
	curAddrs    [3]string
)

var DataNodeAddresses [3]string = [3]string{dn1Addrs, dn2Addrs, dn3Addrs}

type server struct {
	pb.UnimplementedNameNodeServer
}

func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
	log.Println("Sirviendo Solicitud de Registrar Jugada")

	// Elegir un Datanode y guarda la info
	// creado un DataNodeClient y usando la RPC DataNode.RegistrarJugadas

	// Elegimos un DataNode al Azar
	selDataNode := curAddrs[2]
	election := funcs.RandomInRange(1, 3)
	if election == 1 {
		selDataNode = curAddrs[0]
	} else if election == 2 {
		selDataNode = curAddrs[1]
	}

	// Escribirmos el Registro
	AgregarRegistro(in.JugadasJugador, selDataNode)

	// Almacenar Informacion al DataNode Seleccionado
	AlmacenamientoDataNode(in.JugadasJugador, selDataNode)

	return &pb.RespuestaRegistrarJugadas{NumJugador: &pb.JugadorId{Val: in.JugadasJugador.NumJugador.Val}}, nil
}

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverJugadasNameNode) (*pb.RespuestaDevolverJugadas, error) {
	log.Println("Sirviendo Solicitud de Devolver Jugada")
	var numJugador = in.NumJugador.Val

	// jj = jugadas jugador
	// Funcion "LeerRegistroDeJugadas"
	/*
		toma un numero de jugador y se preparar y solicitar toda la info de tal jugador
		Ocupando el registro para buscarla en los respectivos DataNodes.
	*/
	jj, err := LeerRegistroDeJugadas(numJugador)

	if err == nil {
		fmt.Print("Error en linea 92")
		return nil, err
	}
	return &pb.RespuestaDevolverJugadas{JugadasJugador: jj}, nil
}

// Funciones para REGISTRAR JUGADA

func AgregarRegistro(jugadasEnv *pb.JugadasJugador, selDataNodeAddrs string) error {
	jugador := jugadasEnv.NumJugador.Val
	juego := jugadasEnv.JugadasJuego[0].NumJuego

	jugadorStr := strconv.FormatInt(int64(jugador), 10)
	rondaStr := strconv.FormatInt(int64(juego), 10)

	var linea string = "Jugador_" + jugadorStr + " Etapa_" + rondaStr + " " + selDataNodeAddrs + "\n"

	file, err := os.Open("NameNode/" + nameNodeFile)

	if err != nil {
		fmt.Println(err)
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text() + "\n"

		if linea == line {
			fmt.Print("El archivo y el registro ya existe")
			file.Close()
			return nil
		}
	}
	file.Close()

	file2, err2 := os.OpenFile("NameNode/"+nameNodeFile, os.O_APPEND|os.O_WRONLY, 0777)

	if err != nil {
		fmt.Println(err)
		return err2
	}
	defer file2.Close()

	fmt.Printf("Linea registro: %v\n", linea)

	_, err3 := file2.WriteString(linea)

	if err3 != nil {
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

	if err != nil {
		log.Fatalf("Error al registrar jugadas en el DataNode %v: err: %v\n", selDataNodeAddrs, err)
		return err
	}

	fmt.Printf("Se registró la jugada del Jugador: %v\n", r.NumJugador.Val)
	return nil
}

// Funciones para DEVOLER JUGADAS

func LeerRegistroDeJugadas(numjugador int32) (*pb.JugadasJugador, error) {

	// Leemos el archivo de registro
	file, ferr := os.Open("NameNode/" + nameNodeFile)
	if ferr != nil {
		log.Fatalf("Error al abrir archivo - linea 183: err: %v\n", ferr)
		return nil, ferr
	}

	// objeto scannerpara leer linea por linea
	scanner := bufio.NewScanner(file)

	// Creamos una respuesta vacía, pero le agregamos el Numero del Jugador
	// DevolverJugadas devuelve una RespuestaDevolverJugadas
	// JugadasJugador
	// Numero del Jugador (tipo pb.JugadorId -> Val: int32 que es en número del jugador)
	// Arreglo de JugadasJuego
	// JugadaJuego es:
	// Numero de Juego (tipo pb.JUEGO)
	// Arreglo de Jugadas (tipo pb.Jugada -> Val: int32 que es la jugada)

	res := &pb.JugadasJugador{}
	res.NumJugador = &pb.JugadorId{Val: numjugador}

	/*
		dirEtapasMap: Mapa que tiene como llave la dirección de un DataNode (string)
					y como valor un arreglo de strings, que corresponde a números de Etapas

					Ejemplo
					"localhost:50059":["1", "2", ...,"x"]
					"localhost:50058":["1", "3", ..., "y"]

		Recorremos el archivo de registro buscando las lineas que contengan el jugador que buscamos, es decir "Jugador_NumeroJugador".
		Para hacer eso hacemos Split(line, " "), quedando: items[0]: Jugador_x, items[1]: Etapa_y, items[2]: ip_datanode.
		Si la linea corresponde al jugador que buscamos, guardamos como llave su ip_datanode y agregamos el "y" de "Etapa_y" al
		arreglo de string de etapas.
		Si tenemos más de una etapa para el mismo jugador en el mismo datanode, lo agregamos al arreglo de etapas en la misma dirección (llave)
	*/
	dirEtapasMap := make(map[string][]string)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, " ")
		//split: items[0] = Jugador_numero - items[1] = Etapa_numero - items[2] = ip_datanode

		if items[0] == "Jugador_"+funcs.FormatInt32(numjugador) {
			fmt.Println(items)
			_, ok := dirEtapasMap[items[2]]
			if ok {
				dirEtapasMap[items[2]] = append(dirEtapasMap[items[2]], strings.Split(items[1], "_")[1])
			} else {
				dirEtapasMap[items[2]] = []string{strings.Split(items[1], "_")[1]}
			}
		}
	}

	// Recorremos el Map (o Dictionary) creado anteriormente
	// Por cada dirección, nos conectamos al respectivo DataNode y le enviamos las etapas que sabemos que están ahí (1 o más)
	for dialDir, etapas := range dirEtapasMap {

		// Creamos la conexión
		conn, err := grpc.Dial(dialDir, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
			return nil, err
		}
		defer conn.Close()

		// Creamos el DataNodeClient, que deberían estar escuchando
		dc := pb.NewDataNodeClient(conn)

		// Realizamos el Procedimiento Remoto
		// Recibe un pb.SolicitudDevolverJugadasDataNode
		// NumJugador -> tipo pb.JugadorId -> Val: int32 -> número del jugador
		// Etapas: Arreglo de strings, que contiene los números de las etapas (["1", "2", ... "x"])
		r, err := dc.DevolverJugadas(context.Background(),
			&pb.SolicitudDevolverJugadasDataNode{
				NumJugador: &pb.JugadorId{Val: numjugador},
				Etapas:     etapas,
			})

		if err != nil {
			log.Fatalf("Error: %v\n", err)
			return nil, err
		}
		// Nos devuelve la respuesta "r" que es tipo pb.RespuestaDevolverJugadas
		// r-> JugadasJuego
		// NumJugador: Numero del Jugador
		// Arreglo de pb.JugadasJuego
		// NumJuego, tipo pb.JUEGO
		// Arreglo Jugadas -> tipo pb.Jugada -> Val:int32 que corresponde a la jugada

		// a la respuesta plantilla "res", le agregamos el r.JugadasJugador.JugadasJuego recibido
		res.JugadasJuego = append(res.JugadasJuego, r.JugadasJugador.JugadasJuego...)
	}
	//r.JugadasJugador.NumJugador.GetVal()
	//r.JugadasJugador.JugadasJuego[0].NumJuego
	//r.JugadasJugador.JugadasJuego[0].Jugadas[i].Val

	// Devolvemos el JugadasJugador obtenido
	// Volvemos a "DevolverJugadas", donde a un objeto pb.RespuestaDevolverJugadas, le pasaremos las JugadasJugador
	return res, nil
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
