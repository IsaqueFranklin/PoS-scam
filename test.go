package test

import (
	"fmt"
	"time"
)

func sender(out chan<- int) {
	for i := 1; i <= 5; i++ {
		fmt.Println("Enviando:", i)
		out <- i // Enviando o número para o canal
		time.Sleep(time.Second) // Espera um segundo antes de enviar o próximo número
	}
	close(out) // Fechando o canal após enviar todos os números
}

func receiver(in <-chan int, done chan<- bool) {
	for num := range in { // Recebendo os números do canal até que o canal seja fechado
		fmt.Println("Recebido:", num)
	}
	done <- true // Enviando sinal para o canal 'done' indicando que a recepção foi concluída
}

func main() {
	ch := make(chan int) // Canal para enviar números da goroutine 'sender' para a goroutine 'receiver'
	done := make(chan bool) // Canal para sinalizar que a recepção foi concluída

	go sender(ch) // Iniciando a goroutine 'sender' para enviar números para o canal
	go receiver(ch, done) // Iniciando a goroutine 'receiver' para receber números do canal

	<-done // Esperando até que a recepção seja concluída
	fmt.Println("Comunicação concluída.")
}
