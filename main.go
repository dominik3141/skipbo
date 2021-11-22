package main

import (
    "fmt"
    "net"
    "encoding/json"
    "math/rand"
)

type table = [4]([]int)
type storage = [4]([]int)
type stack = []int
type hand = []int

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


func main() {
    numOfPlayers := 2

    // initialize the game
    var players [numOfPlayers]Player
    for i := 0; i < numOfPlayers; i++ {
        players[i] = newPlayer()
    }

    // wait for players to connect...
    initPlayers(numOfPlayers, players)


}

func initPlayers(numOfPlayers int, players []Player) {
    for i := 0; i < numOfPlayers; i++ {
        ln, err := net.Listen("tcp", ":8080")
        if err != nil {
            panic(err)
        }
        
        conn, err := ln.Accept()
        if err != nil {
            panic(err)
        }

        conn.Write(json.Marshall(players[i]))
        

    }

}

func getCards(num int, cards Cards) []int {
    // get 'num' random cards from Game.cards
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


