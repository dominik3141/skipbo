package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
)

type table = [4]([]int)
type storage = [4]([]int)
type hand = []int

type Stack struct {
	cards   []int
	counter int
}

type Cards struct {
	cards   []int
	counter int
}

type Game struct {
	Table        table
	Storage      []storage // one storage for each player
	VisStack     []int
	cards        Cards
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

const numOfPlayers = 2
const numOfCards = 20
const maxMoves = 100 // not supposed to be a real constrained but to prevent an infinite loop

func main() {
	// initialize the game
	cards := NewCards()
	fmt.Printf("cards: %v\n", cards)
	players := make([]Player, numOfPlayers)
	for i := 0; i < numOfPlayers; i++ {
		players[i] = newPlayer(&cards, i)
		fmt.Printf("player %v: %v\n", i, players[i])
	}

	// wait for players to connect...
	fmt.Println("Players created. Now we are waiting for them to connect.")
	conns := initPlayers()
	fmt.Printf("Connections to players established: %v\n", conns)
	closeConns := func(conns []net.Conn) {
		for _, conn := range conns {
			conn.Close()
		}
	}
	defer closeConns(conns)

	// create initial game
	game := newGame(players, cards)

	// play
	var move Move
	for movNr := 0; movNr < maxMoves; movNr++ {
		fmt.Printf("game: %v\n", game)
		sendGameAndPlayers(&game, players, conns)
		fmt.Printf("Done sending game and players info to players\n")

		fmt.Printf("Waiting for move by player %v\n", game.Turn)
		waitForMove(conns[game.Turn], &move)
		fmt.Printf("Move by player %v: %v\n", game.Turn, move)
		checkAndExecMove(&game, players, move)

		exit := checkIfEnd(&game, players)
		if exit {
			return
		}

		game.Turn = (game.Turn + 1) % (numOfPlayers)
	}
}

func checkAndExecMove(game *Game, players []Player, move Move) {
	player := players[game.Turn]
	switch move.KindOfMove {
	case 1: // Hand -> Table
		heapDst := game.Table[move.Dst]
		heapSrc := player.Hand[move.Src]
		heapDst[len(heapDst)] = heapSrc
		// player.Hand
		// ...
		fmt.Println("case 1")
	case 2: // Stack -> Table
		// ...
		fmt.Println("case 2")
	case 3: // Hand -> Storage
		// ...
		fmt.Println("case 3")
	}
}

func checkIfEnd(game *Game, players []Player) bool {
	if (players[game.Turn]).stack.counter == numOfCards {
		return true
	}
	return false
}

func waitForMove(conn net.Conn, moveP *Move) {
	buffer := make([]byte, 0, 1000) // this buffer could probably be much smaller
	for {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n != 0 {
			buffer = buffer[:n] // why is this empty?

			fmt.Printf("Buffer: %s\n", buffer)

			err = json.Unmarshal(buffer, moveP)
			if err != nil && err != io.EOF {
				panic(err)
			}

			fmt.Printf("Received move: %v \n", *moveP)
			return
		}
	}
}

func sendGameAndPlayers(game *Game, players []Player, conns [](net.Conn)) {
	for i, conn := range conns {
		strGame, err := json.Marshal(game)
		if err != nil {
			panic(err)
		}
		strPlayer, err := json.Marshal(players[i])
		if err != nil {
			panic(err)
		}
		conn.Write(strGame)
		conn.Write(strPlayer)
	}
}

func newGame(players []Player, cards Cards) Game {
	var game Game

	game.NumOfPlayers = numOfPlayers
	game.NumOfCards = numOfCards

	game.cards = cards

	for i := 0; i < 4; i++ {
		game.Table[i] = make([]int, 0, 12)
	}

	storages := make([]storage, numOfPlayers)
	for i := 0; i < numOfPlayers; i++ {
		var storage storage
		for j := 0; j < 4; j++ {
			storage[j] = make([]int, 0, 30)

		}
		storages[i] = storage
	}
	game.Storage = storages

	visStack := make([]int, numOfPlayers)
	for i := 0; i < numOfPlayers; i++ {
		pStack := (players[i]).stack
		visStack[i] = pStack.cards[pStack.counter]
	}
	game.VisStack = visStack

	getTurn := func(ints []int) int {
		var max int
		var PlayerIdMax int
		for PlayerId, k := range ints {
			if k > max {
				max = k
				PlayerIdMax = PlayerId
			}
		}
		return PlayerIdMax
	}

	game.Turn = getTurn(visStack)

	return game
}

func initPlayers() [](net.Conn) {
	conns := make([](net.Conn), numOfPlayers)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for i := 0; i < numOfPlayers; i++ {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		conns[i] = conn
	}

	return conns
}

//
// helper functions:
//

func newPlayer(cards *(Cards), id int) Player {
	hand := getCards(5, cards)
	stack := Stack{cards: getCards(numOfCards, cards),
		counter: 0}

	player := Player{
		stack: stack,
		Hand:  hand,
		ID:    id,
	}
	return player
}

func getCards(num int, cardsP *(Cards)) []int {
	ret := make([]int, num)
	for i := 0; i < num; i++ {
		ret[i] = (*cardsP).cards[(*cardsP).counter+i]
	}
	(*cardsP).counter += num
	return ret
}

func NewCards() Cards {
	constAppend := func(c int, n int, slice []int) []int {
		for i := 0; i < n; i++ {
			slice = append(slice, c)
		}
		return slice
	}

	newSkipboSorted := func() []int {
		array := make([]int, 0, 162)
		for i := 0; i < 12; i++ {
			array = constAppend(i+1, 12, array)
		}
		array = constAppend(13, 18, array)
		return array
	}

	cards := newSkipboSorted()
	var rCards Cards
	rCards.cards = make([]int, 162)
	rand.Seed(2606) // not a correct random seed yet!
	for i := 0; i < 162; i++ {
		k := rand.Intn(162)
		if cards[k] != 0 {
			rCards.cards[i] = cards[k]
			cards[k] = 0
		} else {
			i = i - 1
		}
	}
	return rCards
}
