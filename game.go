package main

import (
	"sort"
	"time"

	// "sync"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
)

const (
	COLOR_RED    byte = 'r'
	COLOR_PURPLE byte = 'p'
	COLOR_GREEN  byte = 'g'

	SHAPE_RHOMBUS byte = 'r'
	SHAPE_OVAL    byte = 'o'
	SHAPE_WORM    byte = 'w'

	FILLING_FULL  byte = 'f'
	FILLING_HALF  byte = 'h'
	FILLING_EMPTY byte = 'e'
)

var (
	arrayColor = [3]byte{
		COLOR_RED,
		COLOR_PURPLE,
		COLOR_GREEN,
	}

	mapColor = map[byte]byte{
		byte(COLOR_RED):    COLOR_RED,
		byte(COLOR_PURPLE): COLOR_PURPLE,
		byte(COLOR_GREEN):  COLOR_GREEN,
	}

	arrayShape = [3]byte{
		SHAPE_RHOMBUS,
		SHAPE_OVAL,
		SHAPE_WORM,
	}

	mapShape = map[byte]byte{
		byte(SHAPE_RHOMBUS): SHAPE_RHOMBUS,
		byte(SHAPE_OVAL):    SHAPE_OVAL,
		byte(SHAPE_WORM):    SHAPE_WORM,
	}

	arrayFilling = [3]byte{
		FILLING_FULL,
		FILLING_HALF,
		FILLING_EMPTY,
	}

	mapFilling = map[byte]byte{
		byte(FILLING_FULL):  FILLING_FULL,
		byte(FILLING_HALF):  FILLING_HALF,
		byte(FILLING_EMPTY): FILLING_EMPTY,
	}
)

type Card [4]byte

func (c Card) String() string {
	return string(c[:])
}

func ParseCard(inp string) (Card, error) {
	if len(inp) != 4 {
		return Card{}, ErrBadBody
	}

	color, ok := mapColor[inp[0]]
	if !ok {
		return Card{}, ErrBadBody
	}

	shape, ok := mapShape[inp[1]]

	if !ok {
		return Card{}, ErrBadBody
	}

	filling, ok := mapFilling[inp[2]]

	if !ok {
		return Card{}, ErrBadBody
	}

	if inp[3] > '9' || inp[3] < '0' {
		return Card{}, ErrBadBody
	}

	return Card{color, shape, filling, inp[3]}, nil
}

type User struct {
	ID string `json:"id"`
	// Lock   sync.Mutex      `json:"-"`
	Name   string          `json:"username"`
	Wins   int             `json:"wins"`
	Losses int             `json:"losses"`
	Token  string          `json:"-"`
	Conn   *websocket.Conn `json:"-"`
	Player *Player         `json:"-"`
}

type Player struct {
	// Lock sync.Mutex `json:"-"`
	User *User
	Game *Game `json:"-"`
	// -1 for spectator
	SetsWon int
	Silent  bool
}

type GameState int

const (
	GS_NONE GameState = iota
	GS_WAITING
	GS_PLAYING
)

type Game struct {
	// Lock    *sync.Mutex
	LastCall time.Time `json:"-"`
	ID       string    `json:"id"`
	Deck     []*Card   `json:"-"`
	Owner    *User
	Board    []*Card
	State    GameState
	Players  map[string]*Player
}

func NewGame(owner *User) *Game {
	deck := []*Card{}

	for _, c := range arrayColor {
		for _, s := range arrayShape {
			for _, f := range arrayFilling {
				for n := '0'; n < '3'; n++ {
					deck = append(deck, &Card{c, s, f, byte(n)})
				}
			}
		}
	}

	game := &Game{
		ID:    SnowNode.Generate().String(),
		Deck:  deck,
		Board: []*Card{},
		// Lock: &sync.Mutex{},
	}

	for len(game.Board) < 12 || !BoardHasSet(game.Board) {
		game.AddColumn()
	}

	game.Owner = owner
	game.State = GS_WAITING
	game.Players = map[string]*Player{}

	UserMgr.Add(owner)

	game.AddPlayer(owner)

	return game
}

func (g *Game) AddPlayer(usr *User) {
	// g.Lock.Lock()
	if _, ok := g.Players[usr.ID]; !ok {
		won := 0
		if g.State == GS_PLAYING {
			won = -1
		}
		p := &Player{
			User:    usr,
			SetsWon: won,
			Silent:  false,
			Game:    g,
		}
		usr.Player = p
		g.Players[usr.ID] = p
	}
	// g.Lock.Unlock()
}

func (g *Game) CallSet(playerID string, cards []int) error {
	// g.Lock.Lock()

	if time.Since(g.LastCall) < (time.Millisecond * 250) {
		return nil
	}

	if _, ok := g.Players[playerID]; ok && g.Players[playerID].SetsWon == -1 || g.Players[playerID].Silent {
		return ErrIllegalMove
	}

	for _, ind := range cards {
		if len(g.Board) <= ind {
			return ErrIllegalMove
		}
	}

	if len(cards) != 3 {
		return ErrIllegalMove
	}
	if cards[0] == cards[1] || cards[1] == cards[2] || cards[2] == cards[0] {
		return ErrIllegalMove
	}

	if !CorrectSet(g.Board[cards[0]], g.Board[cards[1]], g.Board[cards[2]]) {
		return ErrNotSet
	}

	g.Players[playerID].SetsWon++

	removeMethod := func() {
		sort.Slice(cards, func(i, j int) bool {
			return cards[i] > cards[j]
		})

		for i := 0; i < 3; i++ {
			slices.Delete(g.Board, cards[i], cards[i]+1)
		}
	}

	notAggressive := func() {
		for i := 0; i < 3; i++ {
			g.Board[cards[i]] = g.GrabDeckCard()
		}
	}

	if len(g.Deck) == 0 {
		removeMethod()
	} else {
		if len(g.Board) == 12 {
			notAggressive()
		} else {
			sI := 0

			for i := len(g.Board) - 1; i > len(g.Board)-4; i-- {
				if slices.Contains(cards, i) {
					continue
				}
				g.Board[cards[sI]] = g.Board[i]
				sI++

				if cards[sI] > len(g.Board)-3-1 {
					break
				}
			}
			g.Board = g.Board[:len(g.Board)-3]
		}
		for !BoardHasSet(g.Board) {
			g.AddColumn()
		}
	}

	g.LastCall = time.Now()

	// g.Lock.Unlock()
	return nil
}

// NOTE: NOT A LOCKED FUNCTION, CAREFUL OF RACE CONDITION
func (g *Game) GrabDeckCard() *Card {
	ind := RandInt(0, len(g.Deck))
	card := g.Deck[ind]
	g.Deck = slices.Delete(g.Deck, ind, ind+1)
	return card
}

func (g *Game) AddColumn() {
	// g.Lock.Lock()
	for i := 0; i < 3; i++ {
		card := g.GrabDeckCard()
		g.Board = append(g.Board, card)
	}
	// g.Lock.Unlock()
}

func BoardHasSet(board []*Card) bool {
	for i1, c1 := range board {
		for i2, c2 := range board {
			if i2 == i1 {
				continue
			}
			for i3, c3 := range board {
				if i3 == i2 || i3 == i1 {
					continue
				}
				if CorrectSet(c1, c2, c3) {
					return true
				}
			}
		}
	}
	return false
}

func BoardHasSetIgnoring(board []*Card, blacklist []int) bool {
	for i1, c1 := range board {
		if slices.Contains(blacklist, i1) {
			continue
		}
		for i2, c2 := range board {
			if i2 == i1 {
				continue
			}
			if slices.Contains(blacklist, i2) {
				continue
			}
			for i3, c3 := range board {
				if i3 == i2 || i3 == i1 {
					continue
				}
				if slices.Contains(blacklist, i3) {
					continue
				}
				if CorrectSet(c1, c2, c3) {
					return true
				}
			}
		}
	}
	return false
}

func CorrectSet(c1, c2, c3 *Card) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			if c1[j] == c2[j] && c1[j] != c3[j] {
				return false
			}
		}
		c2, c3, c1 = c1, c2, c3
	}
	return true
}
