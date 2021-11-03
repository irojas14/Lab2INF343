package Funciones

// 	funcs "github.com/irojas14/Lab2INF343/Funciones"

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
)

func RandomInRange(min int32, max int32) int32 {
	return rand.Int31n(max-min+1) + min
}

func CrearArchivoTxt(NombreDelArchivo string) {
	// check if file exists
	var _, err = os.Stat(NombreDelArchivo)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(NombreDelArchivo)
		if isError(err) {
			return
		}
		defer file.Close()
	}

	fmt.Println("File Created Successfully", NombreDelArchivo)
}

func InsertarJugadasDelJugador(NombreDelArchivo string, JugadasDelJugador []int32) {
	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(NombreDelArchivo, os.O_RDWR, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	for _, jugada := range JugadasDelJugador {
		// Hacer algo
		// Write some text line-by-line to file.
		_, err = file.WriteString(FormatInt32(jugada))
		if isError(err) {
			return
		}
	}

	err = file.Sync()
	if isError(err) {
		return
	}
}

func LeerArchivo(NombreDelArchivo string) {
	// Open file for reading.
	var file, err = os.OpenFile(NombreDelArchivo, os.O_RDWR, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	// Read file, line by line
	var text = make([]byte, 1024)
	for {
		_, err = file.Read(text)

		// Break if finally arrived at end of file
		if err == io.EOF {
			break
		}

		// Break if error occured
		if err != nil && err != io.EOF {
			isError(err)
			break
		}
	}

	fmt.Println("Reading from file.")
	fmt.Println(string(text))
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func FormatInt32(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}
