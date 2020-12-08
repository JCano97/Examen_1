package main

import (
	"bufio"        
	"encoding/gob" 
	"fmt"          
	"io"           
	"log"          
	"net"          
	"os"           
	"strconv"      
	"strings"      
	"time"         
)

var listaClientes []Cliente
var listaMensajes []string
var contadorPuerto int = 1

const TAMANIOBUFFER = 65495

type Cliente struct {
	Nickname string
	numeroPuerto int
}

func servidor() { 
	s, err := net.Listen("tcp", ":9999") //PUERTO :9999 PARA CONECTAR AL SERVIDOR
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleClient(c)
	}
}

func handleClient(c net.Conn) {
	var cliente Cliente
	err := gob.NewDecoder(c).Decode(&cliente)
	if err != nil {
		fmt.Println(err)
		return
	}
	cliente.numeroPuerto = contadorPuerto
	listaClientes = append(listaClientes, cliente)
	contadorPuerto++
	fmt.Println("Ingreso al servidor: ", cliente.Nickname)
	cadena := "Ingreso al servido: " + cliente.Nickname + "\n"
	respaldar(cadena)
}

func recibirMensaje() { //PUERTO :9998 PARA RECIBIR MENSAJE DEL CLIENTE
	s, err := net.Listen("tcp", ":9998")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var mensaje string
		err = gob.NewDecoder(c).Decode(&mensaje)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(mensaje)
		cadena := mensaje + "\n"
		respaldar(cadena)
		for i := 0; i < len(listaClientes); i++ {
			puerto := ":" + strconv.Itoa(listaClientes[i].numeroPuerto)
			go reenviarMensajeClientes(c, mensaje, puerto) //se reenvia el mensaje a todos los clientes
		}
	}
}

func reenviarMensajeClientes(c net.Conn, mensaje string, puerto string) { 
	c, err := net.Dial("tcp", puerto) //PUERTO UNICO DEL CLIENTE PARA ENVIARLE EL MENSAJE 
	if err != nil {
		return
	}
	err = gob.NewEncoder(c).Encode(mensaje)
	if err != nil {
		return
	}
	c.Close()
}

func terminarCliente() { //PUERTO :9997 PARA TERMINA UN CLIENTE
	s, err := net.Listen("tcp", ":9997")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var clienteATerminar string
		err = gob.NewDecoder(c).Decode(&clienteATerminar)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(clienteATerminar + " salio del servidor. . .")
		cadena := clienteATerminar + " salio del servidor. . ." + "\n"
		respaldar(cadena)
		for i := 0; i < len(listaClientes); i++ {
			if listaClientes[i].Nickname == clienteATerminar {
				copy(listaClientes[i:], listaClientes[i+1:])
				listaClientes[len(listaClientes)-1] = Cliente{}
				listaClientes = listaClientes[:len(listaClientes)-1]
			}
		}
	}
}

func asignarPuerto() { //PUERTO :9996 PARA ASIGNA UN PUERTO UNICO A UN CLIENTE
	s, err := net.Listen("tcp", ":9996")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		numero := contadorPuerto
		err = gob.NewEncoder(c).Encode(numero)
	}
}

func conexionRecibirArchivo() {
	s, err := net.Listen("tcp", ":9995") //PUERTO :9995 PARA RECIBIR ARCHIVOS DE CLIENTE
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go recibirArchivo(c)
	}
}

func recibirArchivo(c net.Conn) {
	defer c.Close()

	buferNombreArchivo := make([]byte, 64)//Se crean dos bufers para la creacion del archivo.
	buferTamanioArchivo := make([]byte, 10)

	c.Read(buferTamanioArchivo)//se lee el tamaño del archivo
	tamanioArchivo, _ := strconv.ParseInt(strings.Trim(string(buferTamanioArchivo), ":"), 10, 64)//se guarda el tamaño del archivo

	c.Read(buferNombreArchivo)//se lee el nombre del archivo
	nombreArchivo := strings.Trim(string(buferNombreArchivo), ":")// se guarda el nombre del archivo

	archivo, err := os.Create(nombreArchivo)//se crea un archivo
	if err != nil {
		return
	}
	defer archivo.Close()

	var bytesLeidos int64

	for {//se lee toda la información del archivo y se guarda en el archivo creado
		if (tamanioArchivo - bytesLeidos) < TAMANIOBUFFER {
			io.CopyN(archivo, c, (tamanioArchivo - bytesLeidos))
			c.Read(make([]byte, (bytesLeidos+TAMANIOBUFFER)-tamanioArchivo))
			break
		}
		io.CopyN(archivo, c, TAMANIOBUFFER)
		bytesLeidos += TAMANIOBUFFER
	}
	fmt.Println("Se agrego un nuevo archivo al servidor: ", nombreArchivo)
	cadena := "Se agrego un nuevo archivo al servidor: " + nombreArchivo + "\n"
	respaldar(cadena)
}

func recibirSolicitudArchivo() {
	s, err := net.Listen("tcp", ":9994") //PUERTO :9994 PARA RECIBIR SOLICITUD DE UN ARCHIVO
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		var archivo string
		err = gob.NewDecoder(c).Decode(&archivo)
		if err != nil {
			fmt.Println(err)
			return
		}
		cadena := "Se solicito este archivo del servidor: " + archivo + "\n"
		respaldar(cadena)
		go conexionEnviarArchivo(archivo)
	}
}

func conexionEnviarArchivo(archivo string) {
	s, err := net.Listen("tcp", ":9993") //PUERTO :9993 PARA ENVIO DEL ARCHIVO
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()
	c, err := s.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	go enviarArchivo(c, archivo)
}

func enviarArchivo(c net.Conn, archivo string) {
	auxArchivo, err := os.Open(archivo) //se abre el archivo
	if err != nil {
		fmt.Println(err)
		return
	}

	datosArchivo, err := auxArchivo.Stat() //se lee la informacion del archivo
	if err != nil {
		fmt.Println(err)
		return
	}

	tamanioArchivo := complementarCadena(strconv.FormatInt(datosArchivo.Size(), 10), 10)//se guarda el tamaño del archivo (para enviarlo)

	nombreArchivo := complementarCadena(datosArchivo.Name(), 64)// se guarda el nombre del archivo (para enviarlo)

	c.Write([]byte(tamanioArchivo))//Se escribe la informacion almacenada en las strings creadas previamente
	c.Write([]byte(nombreArchivo))

	lectorBufer := make([]byte, TAMANIOBUFFER)//Se debe crear un bufer, de donde el servidor leera por paquetes.

	for {
		_, err = auxArchivo.Read(lectorBufer)//se lee cada parte del archivo y es enviada
		if err == io.EOF {
			break
		}
		c.Write(lectorBufer)
	}

	cadena := "Se envio un archivo: " + archivo + "\n"
	respaldar(cadena)

	fmt.Println("Se envio un archivo: " + archivo)

	c.Close()
	return
}

func complementarCadena(cadena string, longitud int) string {
	for {
		longitudCadena := len(cadena)
		if longitudCadena < longitud {
			cadena = cadena + ":"
			continue
		}
		break
	}
	return cadena
}

func pausa() { 
	var salto string
	fmt.Scanln(&salto)
	fmt.Println("Presionce ENTER para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func inicializarRespaldo() {
	err := os.Remove("respaldo.txt")
	if err != nil {
		fmt.Println(err)
	}
	tiempoActual := time.Now()
	cadenaTiempo := tiempoActual.Format("2006-01-02 15:04:05")
	cadena := "--------------------------------\n" + cadenaTiempo + " -> Servidor Iniciado\n"
	listaMensajes = append(listaMensajes, cadena)
}

func respaldar(cadena string) {
	tiempoActual := time.Now()
	cadenaTiempo := tiempoActual.Format("2006-01-02 15:04:05")
	cadenaLog := cadenaTiempo + " - " + cadena
	listaMensajes = append(listaMensajes, cadenaLog)
}

func menu() {
	time.Sleep(time.Millisecond)
	continuar := true
	var opcion int
	inicializarRespaldo()
	for continuar != false {
		fmt.Println("-----------------------------")
		fmt.Println("SERVIDOR")
		fmt.Println("1) Chatroom")
		fmt.Println("2) Respaldar mensajes/archivos")
		fmt.Println("3) Terminar servidor")
		fmt.Print("Opción: ")
		fmt.Scan(&opcion)
		switch opcion {
		case 1:
			fmt.Println("Modo chatroom. . .")
			pausa()
		case 2:
			archivo, err := os.OpenFile("respaldo.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("Error al crear el archivo: %s", err)
			}
			datawriter := bufio.NewWriter(archivo)
			for _, data := range listaMensajes {
				_, _ = datawriter.WriteString(data)
			}
			datawriter.Flush()
			archivo.Close()
			fmt.Println("Respaldado. . .")
		case 3:
			fmt.Println("Terminando el servidor. . .")
			cadena := "Servidor Terminado"
			respaldar(cadena)
			continuar = false
		default:
			fmt.Println("Opción no valida")
		}
	}
}

func main() {
	go servidor()
	go recibirMensaje() 
	go terminarCliente() 
	go asignarPuerto() 
	go conexionRecibirArchivo()
	go recibirSolicitudArchivo()
	menu()
}