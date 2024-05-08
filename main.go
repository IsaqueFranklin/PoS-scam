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

//announcements broadcasts winning validator to all nodes.
var announcements = make(chan string)

var mutex = &sync.Mutex{}
//is a standard variable that allows us to control reads/writes and prevent data races

//validators keep track of open validators and balances
var validators = make(map[string]int)


//calculateHash is a simple sha256 hashing function
func calculateHash(s string) string {
  h := sha256.New()
  h.Write([]byte(s))
  hashed := h.sum(nil)
  
  return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all the block information concatenated and sha256 hashed.
func calculateBlockHash(block Block) string {
  record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
  return calculateHash(record)
}

//generateBlock cria um novo bloco usando o hash do bloco passado.
func generateBlock(oldBlock Block, BPM int, address string) (Block, error) {
  
  var newBlock Block

  t := time.Now()

  newBlock.Index = oldBlock.Index + 1
  newBlock.Timestamp = t.String()
  newBlock.BPM = BPM
  newBlock.PrevHash = oldBlock.Hash
  newBlock.Hash = calculateBlockHash(newBlock)
  newBlock.Validator = address
  //The validator field is to know the winning node that forged the block.

  return newBlock, nil
}

//isBlockValid certifica que o bloco é válido checando o index e comparando o hash atual com o hash anterior.
func isBlockValid(newBlock, oldBlock Block) bool {
  if oldBlock.Index+1 != newBlock.Index {
    return false
  }

  if oldBlock.Hash != newBlock.PrevHash {
    return false
  }

  if calculateBlockHash(newBlock) != newBlock.Hash {
    return false
  }

  return true
}

func handleConn(conn net.Conn){
  defer conn.Close()

  go func(){
    for {
      msg := <-announcements
      io.WriteString(conn, msg)
    } 
  }()

  //validator address
  var adress string

  //allow user to allocate number of tokens to stake
  //the grater the number of tokens, the greater chance for forging a new block.
  io.WriteString(conn, "Enter token balance:")
  scanBalance := bufio.NewScanner(conn)

  for scanBalance.Scan() {
    balance, err := strconv.Atoi(scanBalance.Text())
    if err != nil {
      log.Printf("%v not a number: %v", scanBalance.Text(), err)
      return
    }

    t := time.Now()
    address = calculateHash(t.String())
    validators[address] = balance
    fmt.Println(validators)
    break
  }

  io.WriteString(conn, "\nEnter a new BPM:")

  scanBPM := bufio.NewScanner(conn)

  go func() {
    for {
      //take in BPM stdin and add it to blockchain after conducting necessaru validation.
      for scanBPM.Scan() {
        bpm, err := strconv.Atoi(scanBPM.Text())
        //if a malicious party tries to mutate the chain with a bad input, delete them as a validator and they lose their staked coins.
        if err!= nil {
          log.Printf("%v not a number: %v", scanBPM.Text(), err)
          delete(validators, address)
          conn.Close()
        }

        mutex.Lock()
        oldLastIndex := Blockchain[len(Blockchain)-1]
        mutex.Unlock()

        //create newBlock for consideration to be forged
        newBlock, err := generateBlock(oldLastIndex, bpm, address)
        if err != nil {
          log.Println(err)
          continue
        }
        if isBlockValid(newBlock, oldLastIndex){
          candidateBlocks <- newBlock
        }
        io.WriteString(conn, "\nEnter a new BPM: ")
      }
    }
  }()

  //simulate receiving broadcast 
  for {
    time.Sleep(time.Minute)
    mutex.Lock()
    output, err := json.Marshal(Blockchain)
    mutex.Unlock()
    if err != nil {
      log.Fatal(err)
    }
    io.WriteString(conn, string(output)+"\n")
  } 

}

//pickWinner creates a lottery pool of validators and chooses the validator who gets to forge a block to blockchain by random selecting from the pool, weighted by amount of tokens staked.
func pickWinner(){
  times.Sleep(30 * time.Second)
  mutex.Lock()
  temp := tempBlocks
  mutex.Unlock()

  lotteryPool := []string{}
  if len(temp) > 0 {
    // slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without submitting a block to be forged
    OUTER:
      for _, block := range temp {
        //if alrealdy in lottery pool, skip
        for _, node := range lotteryPool {
          if block.Validator == node {
            continue OUTER
          }
        }

        //lock list of validators to prevent data race 
        mutex.Lock()
        setValidators := validators 
        mutex.Unlock()

        k, ok := setValidators[block.Validator]
        if ok {
          for i := 0; i < k; i++ {
            lotteryPool = append(lotteryPool, block.Validator)
          }
        }
      }

      //randomly pick winner from lotteryPool
      s := rand.NewSource(time.Now().Unix())
      r := rand.New(s)
      lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

      //add block of winner to blockchain and let all the other nodes know
      for _, block := range temp {
        if block.Validator == lotteryWinner {
          mutex.Lock()
          Blockchain = append(Blockchain, block)
          mutex.Unlock()
          for _ = range validators {
            announcements <- "\nwinning validator: " + lotteryWinner + "\n"
          }
          break
        }
      }
  }

  mutex.Lock()
  tempBlocks = []block{}
  mutex.Unlock()
}
//We pick a winner every 30 seconds to give time for each validator to propose a new block.
