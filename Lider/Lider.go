package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	MaxPlayers      = 2
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress     = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

var waitCountMux sync.Mutex
var waitingCount int32 = 0
var CurrentAlivePlayers int32 = 0

var (
	ResponsesCount    int32 = 0
	ResponsesCountMux sync.Mutex
)

var (
	JuegoActual       pb.JUEGO
	gameReadyChan     chan int32 = make(chan int32)
	gameLiderValChan  chan int32 = make(chan int32)
	inGameWaitingChan chan int32 = make(chan int32)
)

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Printf("Jugador Uniéndose - jugadores actuales: %v\n", waitingCount)

	if waitingCount > MaxPlayers {
		return nil
	}

	waitCountMux.Lock()
	waitingCount++
	countCont := waitingCount
	jugNum := &pb.JugadorId{Val: countCont}
	waitCountMux.Unlock()

	fmt.Printf("Jugador Count: %v - jugNum: %v\n", waitingCount, jugNum)

	if jugNum.GetVal() < MaxPlayers {
		res := &pb.RespuestaUnirse{
			MsgTipo:    pb.RespuestaUnirse_Esperar,
			NumJugador: jugNum,
			NumJuego:   pb.JUEGO_None,
			NumRonda:   nil,
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
			return err
		}

	} else if jugNum.GetVal() == MaxPlayers {
		res := &pb.RespuestaUnirse{
			MsgTipo:    pb.RespuestaUnirse_Comenzar,
			NumJugador: jugNum,
			NumJuego:   pb.JUEGO_Luces,
			NumRonda:   &pb.RondaId{Val: 0},
		}
		if err := stream.Send(res); err != nil {
			return err
		}
		CurrentAlivePlayers = res.GetNumJugador().GetVal()
		fmt.Printf("CurrentAlivePlayers: %v\n", CurrentAlivePlayers)
		gameReadyChan <- 0
	}
	return nil
}

func (s *server) EnviarJugada(ctx context.Context, in *pb.SolicitudEnviarJugada) (*pb.RespuestaEnviarJugada, error) {

	fmt.Printf("Procesando una jugada enviada: Responses: %v\n", ResponsesCount)
	estado := pb.ESTADO_Muerto

	ResponsesCountMux.Lock()
	ResponsesCount++
	ResponsesCountMux.Unlock()

	fmt.Printf("Responses Tomada en Cuenta: %v\n", ResponsesCount)
	fmt.Printf("Responses == Current?: %v\n", ResponsesCount == CurrentAlivePlayers)
	if ResponsesCount == CurrentAlivePlayers {
		fmt.Println("RespuestasListas")
		inGameWaitingChan <- 0
	}

	jugadaLider := <-gameLiderValChan

	if in.GetJugadaInfo().NumJuego == pb.JUEGO_Luces {
		fmt.Printf("Valor Líder: %v\n", jugadaLider)
		if jugadaLider >= in.GetJugadaInfo().GetJugada().GetVal() {
			estado = pb.ESTADO_Vivo
		}
	} /* else if (in.GetJugadaInfo().NumJuego == pb.JUEGO_TirarCuerda) {
		fmt.Printf("Valor Líder: %v\n", jugadaLider)
		if (jugadaLider >= in.GetJugadaInfo().GetJugada().GetVal()) {
			estado = pb.ESTADO_Vivo
		}
	}
	*/
	return &pb.RespuestaEnviarJugada{NumJuego: in.GetJugadaInfo().GetNumJuego(), JugadaLider: &pb.Jugada{Val: jugadaLider}, Estado: estado}, nil
}

func VerMonto() {
	dialAddrs := pozoAddress
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
	go Update()
	<-waitc
}

func Update() {
	for {
		if JuegoActual == pb.JUEGO_Luces {
			JuegoLucesWaitForResponses()
		} /*else if (JuegoActual == pb.JUEGO_TirarCuerda) {
			JuegoTirarCuerdaWaitForResponses()
		}*/
	}
}

func JuegoLucesWaitForResponses() {
	<-inGameWaitingChan
	gameLiderValChan <- funcs.RandomInRange(6, 10)
}

/*
func JuegoTirarCuerdaWaitForResponses() {
	<-inGameWaitingChan
	gameLiderValChan<- funcs.RandomInRange(1, 4)
}
*/

var changeStateChannel chan int32 = make(chan int32)

func CambiarEtapa(nEtapa pb.JUEGO) {

	fmt.Printf("EstadoActual: %v - Nuevo: %v\n", JuegoActual, nEtapa)

	// Pre-procesamiento
	if JuegoActual == pb.JUEGO_None {

	} else if JuegoActual == pb.JUEGO_Luces {

	} else if JuegoActual == pb.JUEGO_TirarCuerda {
		/*
			jugadores := []int{1, 2, 3, 4, 5, 6, 7, 8}// acá debería ir el array con los jugadores actuales
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(jugadores), func(i, j int) { jugadores[i], jugadores[j] = jugadores[j], jugadores[i] }) // se desordenan
			var largo_jugadores int = len(jugadores)
			if (largo_jugadores % 2){
				judaroes = jugadores[0:largo_jugadores-1] //elimina el ultimo (despues del shuffle por lo que igual es random)
			}
			var team_1 []int = jugadores[0:largo_judadores/2]
			var team_2 []int = jugadores[largo_jugadores/2:largo_jugadores]
		*/
	} else if JuegoActual == pb.JUEGO_TodoNada {

	} else if JuegoActual == pb.JUEGO_Fin {

	}

	JuegoActual = nEtapa

	// Post-procesamiento
	if JuegoActual == pb.JUEGO_None {

	} else if JuegoActual == pb.JUEGO_Luces {

	} else if JuegoActual == pb.JUEGO_TirarCuerda {

	} else if JuegoActual == pb.JUEGO_TodoNada {

	} else if JuegoActual == pb.JUEGO_Fin {

	}
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
