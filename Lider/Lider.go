package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"os"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	MaxPlayers = 16
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)
var jugCountMux sync.Mutex;
var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Println("Jugador Uniéndose")
	
	jugCountMux.Lock()
	defer jugCountMux.Unlock()

	if (jugadorCount < MaxPlayers) {
		res := &pb.RespuestaUnirse{
			MsgTipo: pb.RespuestaUnirse_Esperar,
			NumJugador: nil,
			NumJuego: pb.JUEGO_None,
			NumRonda: nil,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return nil
}

type player_info struct {
	NumJugador pb.JugadorId
	NumRonda pb.RondaId
	NumJuego pb.JUEGO
}

var players []player_info

func VerMonto() {
	dialAddrs := pozoAddress;
	if len(os.Args) == 2 {
		dialAddrs = "localhost:50051"
	}
	fmt.Printf("Consultando Pozo - Addr: %s", dialAddrs)
	// Set up a connection to the server.
	
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPozoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.VerMonto(ctx, &pb.SolicitudVerMonto{})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Printf("\nEl Monto Acumulado Actual es: %f\n", r.GetMonto())
}

// Para actuaizar el Proto file, correr
// go get -u github.com/irojas14/Lab2INF343/Proto
// O CON
// go get -u github.com/irojas14/Lab2INF343

func main() {
	waitc := make(chan struct{})
	go LiderService()
	<-waitc
}

func LiderService() {
	srvAddr := address
	if len(os.Args) == 2 {
		srvAddr = local
	}

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	s := grpc.NewServer()
	pb.RegisterLiderServer(s, &server{})
	log.Printf("Juego Inicializado: escuchando en %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

