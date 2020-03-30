package main

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
)

var lista_datos_conexiones list.List
var lista_conexiones list.List
var lista_mensajes list.List
var indice uint64 = 1

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
	fmt.Println("Servicio activo")
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

	error := gob.NewDecoder(c).Decode(&Peticion)
	if error != nil {
		fmt.Println(error)
	} else {
		if Peticion.Tipo == "INICIO" {
			lista_conexiones.PushBack(c)

			Peticion.Conexion.Id = indice

			error = gob.NewEncoder(c).Encode(indice)
			if error != nil {
				fmt.Println(error)
			} else {
				lista_datos_conexiones.PushBack(Peticion.Conexion)
				indice++
				fmt.Println("Usuario conectado : ", Peticion.Conexion.Nombre)
			}
		} else if Peticion.Tipo == "FIN" {
			fmt.Println("Usuario desconectado : ", Peticion.Conexion.Nombre)
			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(Conexion).Id == Peticion.Conexion.Id {
					lista_datos_conexiones.Remove(e)
				}
			}
		} else if Peticion.Tipo == "MENSAJE" {
			fmt.Println("Nuevo Mensaje")

			lista_mensajes.PushBack(Peticion.Mensaje)
			gob.NewEncoder(c).Encode(Peticion)

			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(Conexion).Activo == true {
					puerto := "990" + strconv.FormatUint(e.Value.(Conexion).Id, 10)

					c2, error := net.Dial("tcp", ":"+puerto)

					if error != nil {
						fmt.Println(error)
					} else {
						error2 := gob.NewEncoder(c2).Encode(Peticion.Mensaje)
						if error2 != nil {
							fmt.Println(error2)
							lista_datos_conexiones.Remove(e)
						} else {
							fmt.Println("Mensaje enviado")
						}
					}
				}
			}
		} else if Peticion.Tipo == "ARCHIVO" {
			fmt.Println("Nuevo Archivo")

		} else {
			fmt.Println("Petición desconocida")
		}
		c.Close()
	}
}

func main() {
	go server()
	var input string
	fmt.Scanln(&input)
}
