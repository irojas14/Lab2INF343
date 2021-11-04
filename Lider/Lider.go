package main

import (
	"context"
	"fmt"
	"io"
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
	MaxPlayers      = 5
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress     = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

var waitCountMux sync.Mutex
var waitingCount int32 = 0

var (
	CurrentAlivePlayers int32 = 0
	CurrentRondaMux     sync.Mutex
	CurrentRonda        int32 = 0
)

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

var esperaRespuestas bool = true

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
		CambiarEtapa(pb.JUEGO_Luces)

		for i := 0; i < MaxPlayers-1; i++ {
			gameReadyChan <- 0
		}
	}
	return nil
}

func (s *server) EnviarJugada(stream pb.Lider_EnviarJugadaServer) error {
	var estado pb.ESTADO	
	
	for {		
		fmt.Printf("Procesando una jugada enviada: Responses: %v\n", ResponsesCount)

		fmt.Println("Esperando una Jugada de un Jugador")
		
		in, err := stream.Recv()
		
		fmt.Println("Jugada Recibida")

		if err == io.EOF {
			fmt.Printf("Un error EOF: %v\n", err)
			return nil
		}

		if err != nil {
			fmt.Printf("Un Error no EOF: %v\n", err)
			return err
		}

		// Espacio Lock BEGIN
		ResponsesCountMux.Lock()
		
		ResponsesCount++

		fmt.Printf("Responses Tomada en Cuenta: %v - Jugador: %v\n", ResponsesCount, in.GetNumJugador())
		fmt.Printf("Responses == Current?: %v\n", ResponsesCount == CurrentAlivePlayers)
		
		if ResponsesCount == CurrentAlivePlayers {
			fmt.Printf("RespuestasListas - Jugador Activador: %v\n", in.GetNumJugador())
			ResponsesCountMux.Unlock()		
			
			inGameWaitingChan <- 0
		} else {
			ResponsesCountMux.Unlock()
		}
		// Espacio Lock END

		fmt.Printf("Esperando Una Jugada del Lider: %v\n", in.GetNumJugador())
		jugadaLider := <-gameLiderValChan
		fmt.Printf("Respuestas del Lider recibida - Con valor: %v\n", jugadaLider)

		if in.GetNumJuego() == pb.JUEGO_Luces {
			fmt.Printf("Valor Líder: %v\n", jugadaLider)

			if CurrentRonda < 4 {
				res := &pb.EnvioJugada{
					Tipo:     pb.EnvioJugada_Jugada,
					Rol:      pb.EnvioJugada_Lider,
					NumJuego: pb.JUEGO_Luces,
					Jugada:   &pb.Jugada{Val: jugadaLider},
				}

				if jugadaLider > in.GetJugada().GetVal() {
					estado = pb.ESTADO_Vivo
					res.Estado = estado
					stream.Send(res)
				} else {
					estado = pb.ESTADO_Muerto
					res.Estado = estado
					stream.Send(res)
					CurrentAlivePlayers--
					break
				}

				if CurrentAlivePlayers == 1 {
					fmt.Printf("Ganaste Jugador: %v\n", in.GetNumJugador())

					youReWinner := &pb.EnvioJugada{
						Tipo:     pb.EnvioJugada_Ganador,
						Rol:      pb.EnvioJugada_Lider,
						NumJuego: pb.JUEGO_Luces,
						Jugada:   &pb.Jugada{Val: jugadaLider},
						NumJugador: in.GetNumJugador(),
					}
					stream.Send(youReWinner)
					break
				}

				fmt.Printf("Esperando La Siguiente Ronda: %v - Jugador: %v\n", CurrentRonda, in.GetNumJugador())
				<-inGameWaitingChan
				fmt.Printf("Enviando Aviso de Comienzo de la Siguiente Ronda: %v - Jugador: %v\n", CurrentRonda, in.GetNumJugador())

				nuevaRonda := pb.EnvioJugada{
					Tipo:     pb.EnvioJugada_NuevaRonda,
					NumRonda: &pb.RondaId{Val: CurrentRonda},
				}
				stream.Send(&nuevaRonda)

				fmt.Printf("ESTADO Jugador %v: %v\n", in.NumJugador.Val, estado.String())
			}
		}
	}
	return nil
}

/*
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

	fmt.Println("Esperando")
	jugadaLider := <-gameLiderValChan
	fmt.Println("Respuestas del Lider recibida")

	if in.GetJugadaInfo().GetNumJuego() == pb.JUEGO_Luces {
		fmt.Printf("Valor Líder: %v\n", jugadaLider)
		if jugadaLider > in.GetJugadaInfo().GetJugada().GetVal() {
			estado = pb.ESTADO_Vivo
		} else {
			estado = pb.ESTADO_Muerto
		}
		fmt.Printf("ESTADO Jugador %v: %v\n", in.JugadaInfo.NumJugador.Val, estado.String())
	} /* else if (in.GetJugadaInfo().NumJuego == pb.JUEGO_TirarCuerda) {
		fmt.Printf("Valor Líder: %v\n", jugadaLider)
		if (jugadaLider >= in.GetJugadaInfo().GetJugada().GetVal()) {
			estado = pb.ESTADO_Vivo
		}
	}
*/ /*
	return &pb.RespuestaEnviarJugada{NumJuego: in.GetJugadaInfo().GetNumJuego(), JugadaLider: &pb.Jugada{Val: jugadaLider}, Estado: estado}, nil
}

*/

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
		}
	}
}

func JuegoLucesWaitForResponses() {
	fmt.Printf("En Juego Luces Responses: Responses: %v - Ronda: %v\n", ResponsesCount, CurrentRonda)
	<-inGameWaitingChan
	fmt.Println("En Juego Luces Wait For Responses -> A Crear Respuesta Líder")
	var i int32 = 0
	liderVal := funcs.RandomInRange(6, 10)
	for i = 0; i < CurrentAlivePlayers; i++ {
		gameLiderValChan <- liderVal
	}
	ResponsesCountMux.Lock()
	ResponsesCount = 0
	ResponsesCountMux.Unlock()

	
	fmt.Printf("CurrentAlivePlayers: %v- MaxPlayers: %v\n", CurrentAlivePlayers, MaxPlayers)
	CurrentRondaMux.Lock()
	CurrentRonda++
	if (CurrentRonda >= 4) {
		fmt.Println("Juego Luces Terminado")
		// Matar a los jugadores
		CurrentRonda = 0
	} else {
		esperaRespuestas = true
	}
	CurrentRondaMux.Unlock()

	for i = 0; i < CurrentAlivePlayers; i++ {
		fmt.Printf("Activando: %v\n", i)		
		inGameWaitingChan <- 0
		fmt.Printf("Activado: %v\n", i)		
	}
	fmt.Println("Se Activaron todas las luces que esperaban")
	fmt.Println("Última Línea Juego Luces")
}

/*
func JuegoTirarCuerdaWaitForResponses() {
	<-inGameWaitingChan
	gameLiderValChan<- funcs.RandomInRange(1, 4)
}
*/

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
