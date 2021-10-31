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
	MaxPlayers = 2
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)
var jugCountMux sync.Mutex;
var jugadorCount int32 = 0

var (
	CurrentGame pb.JUEGO
	gameReadyChan chan bool = make(chan bool)
)


type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Printf("Jugador Uniéndose - jugadores actuales: %v\n", jugadorCount)

	if (jugadorCount > MaxPlayers) {
		return nil
	}
	
	jugCountMux.Lock()
	jugadorCount++;
	countCont := jugadorCount
	jugNum := &pb.JugadorId{Val: countCont}
	jugCountMux.Unlock()

	fmt.Printf("Jugador Count: %v - jugNum: %v\n", jugadorCount, jugNum)

	if (jugNum.GetVal() < MaxPlayers) {
		res := &pb.RespuestaUnirse{
			MsgTipo: pb.RespuestaUnirse_Esperar,
			NumJugador: jugNum,
			NumJuego: pb.JUEGO_None,
			NumRonda: nil,
		}
		fmt.Printf("Sending Res: %v\n", res.String())
		if err := stream.Send(res); err != nil {
			return err
		}
		fmt.Println("Esperando más jugadores")

		res.MsgTipo = pb.RespuestaUnirse_Comenzar
		res.NumJuego = pb.JUEGO_Luces
		res.NumRonda = &pb.RondaId{Val: 0}
		
		<-gameReadyChan

		if err := stream.Send(res); err != nil {
			return err;
		}

	} else if (jugNum.GetVal() == MaxPlayers) {
		res := &pb.RespuestaUnirse{
			MsgTipo: pb.RespuestaUnirse_Comenzar,
			NumJugador: jugNum,
			NumJuego: pb.JUEGO_Luces,
			NumRonda: &pb.RondaId{Val: 0},
		}
		if err := stream.Send(res); err != nil {
			return err;
		}
		gameReadyChan<- true
	}
	return nil
}

func (s *server) EnviarJugada(ctx context.Context, in *pb.SolicitudEnviarJugada) (*pb.RespuestaEnviarJugada, error) {

	return &pb.RespuestaEnviarJugada{Estado: pb.ESTADO_Muerto}, nil
}

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

