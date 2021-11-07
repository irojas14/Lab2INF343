ifdef OS
	RM = del /Q
	FixPath = $(subst /,\,$1)
else
	ifeq ($(shell uname), Linux)
		RM = rm -f
		FixPath = $1
	endif
endif

lider: 
	go run Lider/Lider.go
jugadorH: 
	go run Jugador/jugador.go h
jugadorB: 
	go run Jugador/jugador.go
datanode1:
	go run DataNode/datanode.go 1
datanode2:
	go run DataNode/datanode.go 2
datanode3:
	go run DataNode/datanode.go 3
namenode:
	go run NameNode/namenode.go
pozo:
	go run Pozo/pozo.go
clean:
	rm -- **/*.o