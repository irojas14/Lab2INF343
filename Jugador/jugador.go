package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

//fmt.Print(rand.Intn(100), ",")
const (
	port    = ":50052"
	local   = "localhost" + port
	address = "dist149.inf.santiago.usm.cl" + port
)

var (
	ClientNumJugador  pb.JugadorId
	ClientNumRonda    pb.RondaId
	ClientCurrentGame pb.JUEGO
)

var waitc chan string = make(chan string)

func main() {
	dialAddrs := address
	if len(os.Args) == 2 {
		dialAddrs = local
	}
	fmt.Printf("COMENZANDO EL JUGADOR - Addr: %s\n", dialAddrs)
	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewLiderClient(conn)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	stream, err := c.Unirse(context.Background(), &pb.SolicitudUnirse{})
	
	if err != nil {
		log.Fatalf("Error: %v\n", err)
		return;
	}
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("\nError al Recibir: %v - %v\n", c, err)
			break;
		}
		log.Println("\nRespuesta: " + r.String())

		if r.MsgTipo == pb.RespuestaUnirse_Comenzar {
			ClientNumJugador = *r.GetNumJugador()
			ClientNumRonda = *r.GetNumRonda()
			ClientCurrentGame = r.GetNumJuego()
			log.Printf("Valores: %v - %v - %v\n", ClientNumJugador.GetVal(), ClientNumRonda.GetVal(), ClientCurrentGame)
			break;
		}
	}
	if (ClientCurrentGame == pb.JUEGO_Luces) {
		go Luces(c)
	}
	<-waitc
}

// JUEGOS

func Luces(c pb.LiderClient) (error) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	var randval int32 = funcs.RandomInRange(1, 10)
	fmt.Printf("Random Value: %v\n", r1)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.EnviarJugada(ctx, &pb.SolicitudEnviarJugada{
		JugadaInfo: &pb.PaqueteJugada{
			NumJugador: &pb.JugadorId{Val: ClientNumJugador.Val},
			NumJuego:   ClientCurrentGame,
			NumRonda:   &pb.RondaId{Val: ClientNumRonda.Val},
			Jugada:     &pb.Jugada{Val: randval}}})
	
	if err != nil {
		log.Fatalf("Error al jugar luces: %v\n", err)
		return err
	}
	fmt.Printf("Respuesta Jugada - ESTADO: %v", r.GetEstado().String())	
	waitc<- "done"
	return nil
}

// AUXILIAR

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
//genera un random entre 10 y 15.
