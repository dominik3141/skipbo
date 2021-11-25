package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	Table    table
	Storage  []storage // one storage for each player
	VisStack []int
	//	cards        Cards
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
	connP := getConn()
	fmt.Printf("Connection to game master established: %v\n", *connP)
	defer (*connP).Close()
	var game Game
	var me Player
	getInfo(connP, &game, &me)
	fmt.Printf("me: %v, game: %v\n", me, game)

	for movNr := 0; movNr < maxMoves; movNr++ {
		if game.Turn == me.ID {
			move := buildMove(&game, &me)
			fmt.Printf("My move: %v\n", move)
			sendMove(connP, &move)
		}

		getInfo(connP, &game, &me)
		fmt.Printf("movNr: %v, me: %v, game: %v\n", movNr, me, game)
	}
}

func getInfo(connP *net.Conn, gameP *Game, meP *Player) {
	// wait for input on conn:
	buffer := make([]byte, 10000)
	for i := 0; i < 2; i++ {
		// retOnceNotEmpty := func() int {
		// 	for {
		// 		n, err := (*connP).Read(buffer)
		// 		fmt.Printf("Buffer (in loop): %s\n", buffer[:n])
		// 		if err != nil && err != io.EOF {
		// 			panic(err)
		// 		}
		// 		if n > 140 {
		// 			fmt.Println(n)
		// 			return n
		// 		}
		// 	}
		// }

		n, err := (*connP).Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			i = i - 1
			continue
		}
		buffer = buffer[:n]

		// the buffer should be non-emtpy now
		fmt.Printf("Buffer: %s\n", buffer)

		netIn := json.NewDecoder(bytes.NewReader(buffer))
		if i == 0 {
			err := netIn.Decode(gameP)
			fmt.Println(*gameP)
			if err != nil && err != io.EOF {
				panic(err)
			}
		}
		if i == 1 {
			err = netIn.Decode(meP)
			fmt.Println(*meP)
			if err != nil && err != io.EOF {
				panic(err)
			}
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
	fmt.Printf("Wrote %v bytes to connection\n", n)
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
