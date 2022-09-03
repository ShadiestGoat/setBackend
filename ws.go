package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	// "sync"
	// "time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	Subprotocols:     []string{},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {

	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ManagerBase struct {
	// mlock *sync.Mutex
}

type ManagerUser struct {
	ManagerBase
	Users map[string]*User
}

type ManagerGame struct {
	ManagerBase
	Games map[string]*Game
}

// func (mgr *ManagerBase) Lock() {
// 	mgr.mlock.Lock()
// }

// func (mgr *ManagerBase) Unlock() {
// 	mgr.mlock.Unlock()
// }

func (mgr *ManagerUser) Add(user *User) {
	// mgr.Lock()
	mgr.Users[user.ID] = user
	// mgr.Unlock()
}

// Caller is responsible for doing everything such as closing conn & waiting for reconnection
func (mgr *ManagerUser) Remove(user *User) {
	// mgr.Lock()
	delete(mgr.Users, user.ID)
	// mgr.Unlock()
}

func (mgr *ManagerUser) Send(id string, e Event) {
	if u, ok := mgr.Users[id]; ok {
		b, _ := json.Marshal(e)
		if u.Conn != nil {
			u.Conn.WriteMessage(1, b)
		}
	}
}

func (mgr *ManagerUser) SendRaw(id string, msg *websocket.PreparedMessage) {
	u, ok := mgr.Users[id]
	if ok {
		if u.Conn != nil {
			u.Conn.WritePreparedMessage(msg)
		}
	}
}

func (mgr *ManagerGame) Send(id string, e Event) {
	if _, ok := mgr.Games[id]; ok {
		b, _ := json.Marshal(e)
		msg, err := websocket.NewPreparedMessage(1, b)
		PanicIfErr(err)
		mgr.SendRaw(id, msg)
	}

}

func (mgr *ManagerGame) SendRaw(id string, msg *websocket.PreparedMessage) {
	// mgr.Lock()
	g, ok := mgr.Games[id]

	if ok {
		// g.Lock.Lock()
		for _, p := range g.Players {
			if p.User.Conn != nil {
				p.User.Conn.WritePreparedMessage(msg)
			}
		}
		// g.Lock.Unlock()
	}

	// mgr.Unlock()
}

// // TODO: Merge pinging in here
// // TODO: Maybe add a lock?
// // TODO: change the User.Conn rather than re-assign User

var GameMgr = &ManagerGame{}
var UserMgr = &ManagerUser{}

// TODO: Add locks to prevent race conditions
// TODO: Don't panic if error
// TODO: Join as spectator before the game begins

func (u *User) Ping() {
	for {
		if u.Conn != nil {
			u.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Time{})
			
			u.Conn.SetPongHandler(func(appData string) error {

			})
		}
	}
}

func (p *Player) WSBS() {
	for {
		_, msg, err := p.User.Conn.ReadMessage()
		if err != nil {
			// TODO:
			panic(err)
		}
		ev := Event{}
		err = json.Unmarshal(msg, &ev)
		if err != nil {
			// TODO: close the connection
			// break the loop
		}
		switch ev.Event {
		case E_SET:
			game := p.Game
			// game.Lock.Lock()
			cardsChosen := []int{}

			err := json.Unmarshal(ev.Data, &cardsChosen)

			if err != nil {
				panic(err)
			}

			err = game.CallSet(p.User.ID, cardsChosen)

			if err != nil {
				if errors.Is(err, ErrNotSet) {
					p.Silent = true
					GameMgr.Send(game.ID, Event{
						Event: E_STFU,
						Data:  []byte(`"` + p.User.ID + `"`),
					})
				} else {
					UserMgr.Send(p.User.ID, Event{
						Event: E_ERR,
						Data:  []byte(`"` + err.Error() + `"`),
					})
				}
			} else {
				newBoard := []string{}
				for _, c := range game.Board {
					newBoard = append(newBoard, c.String())
				}
				enc, _ := json.Marshal(EventSET{
					PlayerID: p.User.ID,
					Board:    newBoard,
				})

				GameMgr.Send(game.ID, Event{
					Event: E_SET,
					Data:  enc,
				})
			}

			// game.Lock.Unlock()
		case E_START:
			game := p.Game
			if p.User.ID != game.Owner.ID {
				panic("Not owner")
			}
			game.State = GS_PLAYING

			GameMgr.Send(game.ID, Event{
				Event: E_START,
			})
		default:
			if err != nil {
				// TODO: close the connection
				// break the loop
			}
		}
	}
}
