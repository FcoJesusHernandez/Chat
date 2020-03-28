package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

var Con Conexion = Conexion{
	Id:     999,
	Nombre: "",
	Activo: true,
}

var Men Mensaje

type Peticion struct {
	Tipo     string
	Conexion Conexion
	Mensaje  Mensaje
}

type Conexion struct {
	Id     uint64
	Nombre string
	Activo bool
}

type Mensaje struct {
	Id_conexion     uint64
	Nombre_conexion string
	Contenido       string
}

func enviarMensaje() {
	var texto string
	fmt.Println("Mensaje : ")
	fmt.Scanln(&texto)

	Men.Contenido = texto
	Men.Id_conexion = Con.Id
	Men.Nombre_conexion = Con.Nombre

	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {
		//Con.Con = c

		peticion := Peticion{
			Tipo:     "MENSAJE",
			Conexion: Con,
			Mensaje:  Men,
		}

		fmt.Println(peticion)

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func enviarArchivo() {

}

func clienteFin() {
	fmt.Println("Salir")

	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {

		peticion := Peticion{
			Tipo:     "FIN",
			Conexion: Con,
			Mensaje:  Men,
		}

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func clienteInicio() {
	var nombre string
	fmt.Print("Cual es tu nombre : ")
	fmt.Scanln(&nombre)

	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {
		Con.Nombre = nombre
		//Con.Con = c

		peticion := Peticion{
			Tipo:     "INICIO",
			Conexion: Con,
			Mensaje:  Men,
		}

		fmt.Println(peticion)

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		error = gob.NewDecoder(c).Decode(&Con.Id)

		if Con.Id == 999 {
			fmt.Println("Error con inicialización")
		} else {
			fmt.Println("Inicilización exitosa")
		}

		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func menu() uint {
	var opcion = uint(0)
	fmt.Println("Opciónes")
	fmt.Println("1- Enviar Mensaje")
	fmt.Println("2- Enviar Archivo")
	fmt.Println("3- Salir")
	fmt.Scanln(&opcion)
	return uint(opcion)
}

func main() {
	clienteInicio()

	for {
		switch opcion := menu(); {
		case opcion == uint(1):
			enviarMensaje()
		case opcion == uint(2):
			enviarArchivo()
		case opcion == uint(3):
			clienteFin()
			break
		default:
			fmt.Println("Opción no valida")
		}
	}
}
