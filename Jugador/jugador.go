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

func main() {
	dialAddrs := address
	if len(os.Args) == 2 {
		dialAddrs = local
	}
	fmt.Printf("COMENZANDO EL JUGADOR - Addr: %s", dialAddrs)
	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLiderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := c.Unirse(ctx, &pb.SolicitudUnirse{})
	
	if err != nil {
		log.Fatalf("Error: %v\n", err)
		return;
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("\n%v.ListFeatures(_) = _, %v", c, err)
		}
		log.Println("\nRespuesta: %s\n", res)
	}
}

// JUEGOS

func Luces(c pb.LiderClient, ctx context.Context, cancel context.CancelFunc) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	var randval int32 = r1.Int31()
	fmt.Printf("Random Value: %v", r1)

	c.EnviarJugada(ctx, &pb.SolicitudEnviarJugada{
		JugadaInfo: &pb.PaqueteJugada{
			NumJugador: &pb.JugadorId{Val: ClientNumJugador.Val},
			NumJuego:   ClientCurrentGame,
			NumRonda:   &pb.RondaId{Val: ClientNumRonda.Val},
			Jugada:     &pb.Jugada{Val: randval}}})
}

// AUXILIAR

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
