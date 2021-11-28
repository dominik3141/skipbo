package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"
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
const maxMoves = 1000 // not supposed to be a real constrained but to prevent an infinite loop

func main() {
	// wait for players to connect...
	fmt.Println("Players created. Now we are waiting for them to connect.")
	conns := initPlayerConns()
	fmt.Printf("Connections to players established: %v\n", conns)

	var winnCounter [numOfPlayers]int
	logFile, err := os.Create("game.log")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	logW := bufio.NewWriter(logFile)

gameInit:
	// create new cards and create 'numOfPlayers' players
	cards := NewCards()
	fmt.Printf("cards: %v\n", cards)
	players := make([]Player, numOfPlayers)
	for i := 0; i < numOfPlayers; i++ {
		players[i] = newPlayer(&cards, i)
		fmt.Printf("Created player %v: \t %v\n", i, players[i])
	}

	// make sure the connection get closed after the game has ended
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
		// send game and player infos to all players
		fmt.Printf("game: %v\n", game)
		sendGameAndPlayers(&game, players, conns)
		fmt.Printf("Done sending game and players info to players\n")

		// wait for move of player with id game.Turn
		fmt.Printf("Waiting for move by player %v\n", game.Turn)
		waitForMove(conns[game.Turn], &move)
		fmt.Printf("Move nr %v by player %v: %v\n", movNr, game.Turn, move)

		// check and exec move (panics if move is illegal)
		checkAndExecMove(&game, players, move)

		//	check if someone has already won the game
		exit := checkIfEnd(game.Turn, players)
		if exit {
			fmt.Printf("Player %v has won the game!\n", game.Turn)
			winnCounter[game.Turn] += 1
			fmt.Printf("winnCounter: %v\n", winnCounter)
			fmt.Fprintf(logW, "winnCounter: %v, \t moveNr: %v\n", winnCounter, movNr)
			logW.Flush()
			goto gameInit
		}

		// make everything ready for next move/iteration:

		//		determine who has the turn
		prevTurn := game.Turn
		game.Turn = nextTurn(game.Turn, move.KindOfMove, players)

		// 		refill hand cards if player has ended turn and check if some heap on the table is full
		cleanUp(&game, players, prevTurn)

		fmt.Printf("Next turn: player %v\n", game.Turn)
	}
}

func nextTurn(turn int, kindOfMove int, players []Player) int {
	// determine which player is supposed to make the next move.
	if kindOfMove == 3 {
		return (turn + 1) % (numOfPlayers)
	}
	return turn
}

func cleanUp(gameP *Game, players []Player, prevTurn int) {
	// check if some heap on the table is full
	for i := 0; i < 4; i++ {
		if len((*gameP).Table[i]) == 12 {
			(*gameP).Table[i] = (*gameP).Table[i][:0]
		}
	}

	// give new hand cards to players who just ended their turn
	if (*gameP).Turn != prevTurn {
		player := players[prevTurn]
		fmt.Printf("ID: %v \t CleanUp: \t Old Hand: %v\n", player.ID, player.Hand)
		players[prevTurn].Hand = append(players[prevTurn].Hand, getCards(5-len(players[prevTurn].Hand), &(*gameP).cards)...)
		fmt.Printf("ID: %v \t CleanUp: \t New Hand: %v\n", player.ID, players[prevTurn].Hand)
	}

	// give new hand cards to players that have no cards left on their hand
	for id, _ := range players {
		if len(players[id].Hand) == 0 {
			fmt.Printf("ID: %v \t CleanUp: \t Old Hand: %v\n", players[id].ID, players[id].Hand)
			players[id].Hand = append(players[id].Hand, getCards(5-len(players[id].Hand), &(*gameP).cards)...)
			fmt.Printf("ID: %v \t CleanUp: \t New Hand: %v\n", players[id].ID, players[id].Hand)
		}
	}
}

func checkAndExecMove(gameP *Game, players []Player, move Move) {
	deleteFromHand := func(playerP *Player, ind int) {
		// delete the card specified by its index 'ind' from 'playerP's hand and reindex accordingly
		cardsOnHand := len((*playerP).Hand)
		for i := move.Src; i < 5; i++ {
			if i+1 < cardsOnHand {
				(*playerP).Hand[i] = (*playerP).Hand[i+1]
			}
		}
		(*playerP).Hand = (*playerP).Hand[:cardsOnHand-1]
	}
	playerP := &(players[(*gameP).Turn])
	playerID := (*gameP).Turn
	switch move.KindOfMove {
	case 1: // Hand -> Table
		fmt.Println("checkAndExecMove: \t \t case 1")

		// get variables we need
		tableHeapP := &((*gameP).Table[move.Dst])
		handCard := (*playerP).Hand[move.Src]

		// lay down card to table if this is a legit move
		if legit(*tableHeapP, handCard) {
			// (*tableHeapP)[len(*tableHeapP)] = handCard
			*tableHeapP = append(*tableHeapP, handCard)
			fmt.Printf("Move %v by %v is legitimate.\n", move, playerID)
		} else {
			panic("ERROR: illegal move!")
		}

		// delete card from hand
		deleteFromHand(playerP, move.Src)

	case 2: // Stack -> Table
		fmt.Println("checkAndExecMove: \t \t case 2")

		// get visible stack card from stack and increase counter
		card := players[playerID].stack.cards[players[playerID].stack.counter]
		players[playerID].stack.counter += 1

		// lay down card to table if this is a legit move
		if legit((*gameP).Table[move.Dst], card) {
			(*gameP).Table[move.Dst] = append((*gameP).Table[move.Dst], card)
			fmt.Printf("Move %v by %v is legitimate.\n", move, playerID)
		} else {
			panic("ERROR: Illegitimate move!")
		}

		// update visible stack card
		(*gameP).VisStack[playerID] = players[playerID].stack.cards[players[playerID].stack.counter]

	case 3: // Hand -> Storage
		// this move is always legitimate
		fmt.Println("checkAndExecMove: \t \t case 3")
		fmt.Printf("ID: %v \t checkAndExecMove: \t Old Hand: %v\n", playerID, (*playerP).Hand)
		storageDstP := &((*gameP).Storage[playerID][move.Dst])
		HandSrcP := &((*playerP).Hand[move.Src])
		*storageDstP = append(*storageDstP, *HandSrcP)

		deleteFromHand(playerP, move.Src)
		fmt.Printf("ID: %v \t checkAndExecMove: \t New Hand: %v\n", playerID, (*playerP).Hand)
	}
}

func legit(a []int, b int) bool {
	// check if it is legitimate (within the rules of skipbo) to append b to a
	if b == 13 {
		return true
	}
	if len(a) == 0 {
		return b == 1
	}
	if a[len(a)-1] == 13 {
		return legit(a[:len(a)-1], b-1)
	}
	return a[len(a)-1] == b-1
}

func checkIfEnd(playerId int, players []Player) bool {
	// check if someone has already won the game
	return (players[playerId]).stack.counter == numOfCards-1
}

func waitForMove(conn net.Conn, moveP *Move) {
	// waits till the player connected trough 'conn' submits his move
	// the function that saves this move to the Move struct pointed to by 'moveP'

	buffer := make([]byte, 1000) // this buffer could probably be much smaller
	for {
		// time.Sleep(1 * time.Second)

		n, err := conn.Read(buffer)
		// fmt.Println(n, err)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n != 0 {
			buffer = buffer[:n]

			fmt.Printf("waitForMove \t Buffer: \t %s\n", buffer)

			// can not handle more than one move in a single buffer
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
	// sends the game (or at leat all 'public' fields of the game) to each player
	// and send every player his own player variable
	for i, conn := range conns {
		strGame, err := json.Marshal(game)
		if err != nil {
			panic(err)
		}
		conn.Write(strGame)

		strPlayer, err := json.Marshal(players[i])
		if err != nil {
			panic(err)
		}
		conn.Write(strPlayer)
	}
}

func newGame(players []Player, cards Cards) Game {
	// given a slice of players and (the value of) a cards struct, this function returns a game with proper:
	// turn, visStack, cards, numOfPlayers, numOfCards
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

func initPlayerConns() [](net.Conn) {
	// create a connection with each player and return a slice containing the connections
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
	// creates a new player with ID='id' and assignes hand- and stack cards to the player
	hand := getCards(5, cards)
	stack := Stack{cards: getCards(numOfCards, cards),
		counter: 0}

	player := Player{
		ID:    id,
		Hand:  hand,
		stack: stack,
	}
	return player
}

func getCards(num int, cardsP *(Cards)) []int {
	// given an integer 'num' and a pointer to a Cards struct, this function returns the first num cards from the Cards and increases its counter by num
	if (*cardsP).counter > 155 {
		*cardsP = NewCards()
	}
	ret := make([]int, num)
	for i := 0; i < num; i++ {
		ret[i] = (*cardsP).cards[(*cardsP).counter+i]
	}
	(*cardsP).counter += num
	return ret
}

func NewCards() Cards {
	// returns a Cards struct containing:
	//	cards.cards: 	a slice of integers containing all skipbo cards
	// 	cards.counter: 	a counter initialized to zero
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
	rand.Seed(time.Now().UnixNano()) // not a correct random seed yet!
	for i := 0; i < 161; i++ {
		k := rand.Intn(161)
		if cards[k] != 0 {
			rCards.cards[i] = cards[k]
			cards[k] = 0
		} else {
			i = i - 1
		}
	}
	return rCards
}
