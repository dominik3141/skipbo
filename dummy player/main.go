package main

import (
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
	// spawns two players

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

	// get new infos after every move and submit own move if it is my turn
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

func buildMove(gameP *Game, meP *Player) Move {
	// this function is the core of the dummy players strategy

	checkStack := func() (bool, int) {
		// check whether it is possible to lay down the visible stack card to some heap on the table
		// if it is possible, return the index of the correct table heap
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

func getInfo(connP *net.Conn, gameP *Game, meP *Player) {
	netIn := json.NewDecoder(*connP)

	err := netIn.Decode(gameP)
	if err != nil && err != io.EOF {
		panic(err)
	}

	err = netIn.Decode(meP)
	if err != nil && err != io.EOF {
		panic(err)
	}
}

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

//
// helper functions
//

func heads(heaps [4]([]int)) []int {
	// given an array of four int slices, this function returns a slice containing all four first elements (or zero if heap is empty) from these slices
	heads := make([]int, 4)
	for i := 0; i < 4; i++ {
		if len(heaps[i]) == 0 {
			continue
		}
		heads[i] = (heaps[i])[len(heaps[i])-1]
	}
	return heads
}
