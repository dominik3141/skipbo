package main

import (
    "fmt"
    "net"
    "encoding/json"
    "math/rand"
)

type table = [4]([]int)
type storage = [4]([]int)
type hand = []int

type Stack struct {
    cards       []int
    counter     int
}

type Cards struct {
    cards       []int
    counter     int
}

type Game struct {
    Table       table
    Storage     []storage
    VisStack    []int
    cards       cards
    Turn        uint
}

type Player struct {
    ID          uint
    Hand        hand
    stack       stack
}

type Move struct {
    kindOfMove  uint    // i.e. Hand -> Table
    src         uint    // i.e. which card from Hand
    dst         uint    // i.e. which heap to lay down on
}


const numOfPlayers = 2
const numOfCards   = 20


func main() {
    // initialize the game
    cards := NewCards()
    players := make([]Player, numOfPlayers)
    for i := 0; i < numOfPlayers; i++ {
        fmt.Println("Creating player", i)
        players[i] = newPlayer(&cards, i+1)
    }

    // wait for players to connect...
    conns := initPlayers(numOfPlayers, players)

    game := newGame(players, cards)
    for {
        sendGame(&game, players, conns)
        move := waitForMove(conns[game.Turn])
        checkAndExecMove(&game, players, move)
        exit := checkIfEnd(&game, players)
        if exit {
            return
        }
    }
}

func checkAndExecMove(game *Game, players []Player, move Move) {
    player := players[game.Turn]
    switch move.kindOfMove {
    case 1: // Hand -> Table
            heapDst = game.Table[move.Dst]
            heapSrc = player.Hand[move.Src]
            heapDst[len(heapDst)] = heapSrc[len(heapSrc)-1]
            player.Hand
            ...
    case 2: // Stack -> Table
            ...
    case 3: // Hand -> Storage
            ...
    }
}

func checkIfEnd(game *Game, players []Player) Bool {
    if (players[game.Turn]).stack.counter == numOfCards {
        return true
    }
    return false
}

func waitForMove(conn net.Conn) Move {
    var move Move
    buffer := make([]byte, 0, 1000) // this buffer could probably be much smaller
    conn.Read(buffer)
    json.Unmarshall(buffer, move)
    return move
}

func sendGame(game *Game, players []Player, conns [](net.Conn)) {
        turn = (game.Turn + 1) % (numOfPlayers+1)
        if turn == 0 {
            game.Turn = 1
        } else {
            game.Turn = turn
        }
        for i, conn := range conns {
                conn.Write(json.Marshall(game)
                conn.Write(json.Marshall(players[i+1])
        }
}


func newGame(players *([]Player), cards Cards) Game {
    var game Game
    
    game.cards = cards
    
    game.Turn = 1
    
    game.Table = [ make(([]int), 0, 12), make(([]int), 0, 12), make(([]int), 0, 12), make(([]int), 0, 12) ] 

    var storage []int
    for i:=0;i<numOfPlayers;i++ { storage = append(storage, make([]int), 0, 30) }
    game.Storage = storage   

    visStack := make([]int, numOfPlayers)
    for i:=0;i<numOfPlayers;i++ { pStack := (players[i+1]).stack; visStack[i+1] = pStack.cards[pStack.counter]  }
    game.VisStack = visStack

    return game
}

func initPlayers(numOfPlayers int, players []Player) [](net.Conn) {
    var conns [numOfPlayers](net.Conn)
    conns := make([](net.Conn), numOfPlayers)
    for i := 0; i < numOfPlayers; i++ {
        ln, err := net.Listen("tcp", ":8080")
        if err != nil {
            panic(err)
        }
        
        conn, err := ln.Accept()
        if err != nil {
            panic(err)
        }
        conns[i] = conn

        conn.Write(json.Marshall(players[i+1]))
    }
    return conns
}




//
// helper functions:
//


func newPlayer(cards *Cards, id int) Player {
    hand := getCards(5,cards)
    stack := Stack { cards: getCards(numOfCards, cards)
                     counter: 0 }

    player := Player {
                    stack: stack
                    hand: hand
                    id: id
            }
    return player
}

func getCards(num int, cards *Cards) []int {
        ret := make([]int, num)
        for i:= 0; i<num; i++ {
            ret[i] = cards.cards[cards.counter + i]
        }
        cards.counter += num
        return ret
}

func NewCards() []int {
        constA := func (c int, n int) []int {
                        array := make([]int, n)
                        for i:=0; i<n; i++ {
                            array[i] = c
                        }
                     return array
                  }

        newSkipboSorted := func () []int {
                                array := make([]int, 160)
                                for i:=0; i<12; i++ {
                                    array = append(array, *constA(i+1, 12))
                                }
                                array = append(array, *constA(13, 16))
                              return array
                           }

        cards := *newSkipboSorted()
        rCards := make([]int, 0, 160)
        rand.Seed(2606)         // not a correct random seed yet!
        i := 0
        while i < 160 {
                k := rand.Intn(160)
                if cards[k] != 0 {
                    rCards[i] = cards[k]
                    cards[k] = 0
                    i ++
                }
        }
    return rCards
}


