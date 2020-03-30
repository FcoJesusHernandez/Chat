package main

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"net"
)

var lista_datos_conexiones list.List
var lista_conexiones list.List
var lista_mensajes list.List
var indice uint64 = 0

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

func server() {
	s, error := net.Listen("tcp", ":9999") // c = conexion escuchando en el puerto
	if error != nil {
		fmt.Println(error)
		return
	}
	for {
		c, error := s.Accept() //cuando acepte la conexion llamara al manejador
		if error != nil {
			fmt.Println(error)
			continue
		}

		go handleCliente(c)
	}
}

func handleCliente(c net.Conn) {
	var Peticion Peticion
	fmt.Println("ss")
	error := gob.NewDecoder(c).Decode(&Peticion)
	fmt.Println("cos")
	if error != nil {
		fmt.Println(error)
	} else {
		fmt.Println(Peticion)
		if Peticion.Tipo == "INICIO" {
			lista_conexiones.PushBack(c)

			Peticion.Conexion.Id = indice

			error = gob.NewEncoder(c).Encode(indice)
			if error != nil {
				fmt.Println(error)
			} else {
				lista_datos_conexiones.PushBack(Peticion.Conexion)
				indice++
				fmt.Println("Nuevo usuario ", Peticion.Conexion.Nombre)
			}
		} else if Peticion.Tipo == "FIN" {
			fmt.Println("Cliente termino conexion")
			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(*Conexion).Id == Peticion.Conexion.Id {
					e.Value.(*Conexion).Activo = false
				}
			}
		} else if Peticion.Tipo == "MENSAJE" {
			fmt.Println("Nuevo Mensaje")
			fmt.Println(Peticion.Mensaje)

			lista_mensajes.PushBack(Peticion.Mensaje)

			for e := lista_conexiones.Front(); e != nil; e = e.Next() {
				error = gob.NewEncoder(e.Value.(net.Conn)).Encode(Peticion.Mensaje)
				if error != nil {
					fmt.Println(error)
				}
				fmt.Println("mensaje enviado")
			}
		} else if Peticion.Tipo == "ARCHIVO" {
			fmt.Println("Nuevo Archivo")

		} else {
			fmt.Println("Petición desconocida")
		}
		//c.Close()
	}
}

func main() {
	go server()
	var input string
	fmt.Scanln(&input)
}
