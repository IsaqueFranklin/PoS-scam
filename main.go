package main

import (
  "bufio"
  "crypto/sha256"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "math/rand"
  "net"
  "os"
  "strconv"
  "sync"
  "time"

  "github.com/davecgh/go-spew/spew"
  //spew is a convinient package that pretty prints the blockchain to the terminal.
	"github.com/joho/godotenv"
  //godotenv is for reading the .env file.
)

type Block struct {
  Index int
  Timestamp string
  BPM int
  Hash string
  PrevHash string
  Validator string
}

//A Blockchain é uma série de blocos validados.
var Blockchain []Block
var tempBlocks []Block

//candidateBlocks handles incoming blocks for validation.
var candidateBlocks = make(chan Block)
