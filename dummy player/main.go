package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

type table = [4]([]int)
type storage = [4]([]int)
type hand = []int

type Stack struct {
	cards   []int
	counter int
}

// type Cards struct {
// 	cards   []int
// 	counter int
// }

type Game struct {
	Table        table
	Storage      []storage // one storage for each player
	VisStack     []int
	Turn         int
	NumOfPlayers int
	NumOfCards   int
}

type Player struct {
	ID    int
	Hand  hand
	stack Stack
}

type Move struct {
	KindOfMove int // i.e. Hand -> Table
	Src        int // i.e. which card from Hand
	Dst        int // i.e. which heap to lay down on
}

const maxMoves = 100 // not supposed to be a real constrained but to prevent an infinite loop

func main() {
	go MainPlayer()
	MainPlayer()
}

func MainPlayer() {
	// establish connection to game master
	connP := getConn()
	fmt.Printf("Connection to game master established: %v\n", *connP)
	defer (*connP).Close()

	// get infos
	var game Game
	var me Player
	getInfo(connP, &game, &me)
	fmt.Printf("ID: %v \t STATUS: \t  movNr: 0, me: %v, game: %v\n", me.ID, me, game)

	for movNr := 1; movNr < maxMoves+1; movNr++ {
		if game.Turn == me.ID {
			move := buildMove(&game, &me)
			fmt.Printf("ID: %v \t \t My move: %v\n", me.ID, move)
			sendMove(connP, &move)
			fmt.Printf("ID: %v \t \t Done sending move.\n", me.ID)
		}

		getInfo(connP, &game, &me) // this function should also block further execution until some player has send his move to the game master
		fmt.Printf("ID: %v \t STATUS: \t  movNr: %v, me: %v, game: %v\n", me.ID, movNr, me, game)
	}
}

func decodeInfo(netIn *json.Decoder, gameP *Game, meP *Player) {
	smallBuffer := make([]byte, 3)
	netIn.Buffered().Read(smallBuffer)
	switch smallBuffer[2] {
	case byte(0x54):
		err := netIn.Decode(gameP)
		if err != nil && err != io.EOF {
			panic(err)
		}
	case byte(0x49):
		err := netIn.Decode(meP)
		if err != nil && err != io.EOF {
			panic(err)
		}
	}
}

func getInfo(connP *net.Conn, gameP *Game, meP *Player) {
	// wait for input on conn:
	buffer := make([]byte, 10000)

	for {
		time.Sleep(2 * time.Second)
		n, err := (*connP).Read(buffer)
		fmt.Printf("ID: %v \t \t n=%v\n", (*meP).ID, n)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			continue
		}

		// the buffer should be non-emtpy now

		netIn := json.NewDecoder(bytes.NewReader(buffer[:n]))
		decodeInfo(netIn, gameP, meP)

		switch netIn.More() {
		case true:
			decodeInfo(netIn, gameP, meP)
			return
		case false:
			continue

		}
	}
}

func buildMove(gameP *Game, meP *Player) Move {
	checkStack := func() (bool, int) {
		heads := heads((*gameP).Table)
		for i := 0; i < 4; i++ {
			if heads[i] == (*gameP).VisStack[(*meP).ID] {
				return true, i
			}
		}
		return false, 0
	}

	var move Move
	csBool, ind := checkStack()
	if csBool {
		move.KindOfMove = 2
		move.Src = 0
		move.Dst = ind
		return move
	}
	move.KindOfMove = 3
	move.Src = 0
	move.Dst = 0
	return move
}

// func getGameInfo(connP *net.Conn, gameP *Game) {
// 	buffer := make([]byte, 1000)
// 	n, err := (*connP).Read(buffer)
// 	if err != nil && err != io.EOF {
// 		panic(err)
// 	}
// 	buffer = buffer[:n]

// 	err = json.Unmarshal(buffer, gameP)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func getPrivateInfo(connP *net.Conn, me *Player) {
// 	buffer := make([]byte, 1000)
// 	n, err := (*connP).Read(buffer)
// 	if err != nil && err != io.EOF {
// 		panic(err)
// 	}
// 	buffer = buffer[:n]

// 	fmt.Println(buffer)
// 	err = json.Unmarshal(buffer, me)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func getConn() *net.Conn {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	return &conn
}

func sendMove(connP *net.Conn, moveP *Move) {
	bMove, err := json.Marshal(*moveP)
	if err != nil {
		panic(err)
	}

	n, err := (*connP).Write(bMove)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wrote %v bytes to connection: %s\n", n, bMove)
}

// func checkAndExecMove(game *Game, players []Player, move Move) {
// 	player := players[game.Turn]
// 	switch move.KindOfMove {
// 	case 1: // Hand -> Table
// 		heapDst := game.Table[move.Dst]
// 		heapSrc := player.Hand[move.Src]
// 		heapDst[len(heapDst)] = heapSrc
// 		// player.Hand
// 		// ...
// 		fmt.Println("case 1")
// 	case 2: // Stack -> Table
// 		// ...
// 		fmt.Println("case 2")
// 	case 3: // Hand -> Storage
// 		// ...
// 		fmt.Println("case 3")
// 	}
// }

// func checkIfEnd(game *Game, players []Player) bool {
// 	if (players[game.Turn]).stack.counter == numOfCards {
// 		return true
// 	}
// 	return false
// }

//
// helper functions:
//

func heads(heaps [4]([]int)) []int {
	heads := make([]int, 4)
	for i := 0; i < 4; i++ {
		if len(heaps[i]) == 0 {
			continue
		}
		heads[i] = (heaps[i])[len(heaps[i])-1]
	}
	return heads
}
