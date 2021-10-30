package main

import (
	"context"
	"fmt"
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
	numJugador pb.JugadorId
	numRonda   pb.RondaId
)

func main() {
	fmt.Printf("len Args: %v\n", len(os.Args))
	dialAddrs := address
	if len(os.Args) == 2 {
		dialAddrs = local
	}
	fmt.Printf("COMENZANDO EL JUGADOR - Addr: %s", dialAddrs)
	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	fmt.Println("Sgte Linea desde el Dial")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Dial Terminado")

	c := pb.NewLiderClient(conn)

	fmt.Println("Cliente Creado")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Unirse(ctx, &pb.SolicitudUnirse{Solictud: "Quiero Jugar"})
	fmt.Println("Sgte Linea desde el llamado Unirse")

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println("Llamado Remoto Finalizado")
	log.Printf("Bienvenido al Juego: Tu Numero: " + r.GetNumJugador().String() + " Juego: " + r.GetNumJuego().String() + " Numero Ronda: " + r.GetNumRonda().String() + "\n")
}

// JUEGOS

func Luces(c *pb.LiderClient) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	fmt.Printf("Random Value: %v", r1)
}

// AUXILIAR

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
