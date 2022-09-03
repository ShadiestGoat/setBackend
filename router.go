package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4"
)

type ReqContext int

const (
	CTX_USR ReqContext = iota
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawAuth := r.Header.Get("Authorization")
		auth := strings.Split(rawAuth, " ")
		if len(auth) == 0 || len(auth) > 2 {
			WriteErr(w, ErrNotAuthorized)
			return
		}

		id := auth[0]
		var usr *User

		if usrFromMgr, ok := UserMgr.Users[id]; ok {

			if usrFromMgr.Token != "" {
				if len(auth) != 2 || auth[1] != usrFromMgr.Token {
					WriteErr(w, ErrNotAuthorized)
					return
				}
			}

			usr = usrFromMgr
		} else {
			wins, losses, dbToken, name := 0, 0, "", ""
			resp := DBQueryRow(`SELECT wins,losses,token,name FROM users WHERE id=$1`, id)
			if err := resp.Scan(&wins, &losses, &dbToken, &name); err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					panic(err)
				}
				name = id
			}

			if dbToken != "" {
				if len(auth) != 2 || auth[1] != dbToken {
					WriteErr(w, ErrNotAuthorized)
					return
				}
			}

			usr = &User{
				ID:     id,
				Name:   name,
				Wins:   wins,
				Losses: losses,
				Token:  dbToken,
			}
		}

		ctx := context.WithValue(r.Context(), CTX_USR, usr)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func routerAPI() http.Handler {
	r := chi.NewRouter()

	r.Use(httprate.Limit(
		120,           // requests
		3*time.Minute, // per duration
		httprate.WithKeyFuncs(httprate.KeyByEndpoint, func(r *http.Request) (string, error) {
			id := r.Header.Get("auth") // TODO: ip
			if id == "" {
				id = SnowNode.Generate().String()
			}
			return id, nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			WriteErr(w, ErrRateLimit)
		}),
	))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		id := SnowNode.Generate().String()
		RespondString(w, 200, `{"id":"`+id+`"}`)
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Post("/games", func(w http.ResponseWriter, r *http.Request) {
			// usr := r.Context().Value(CTX_USR).(*User)
			// game := NewGame(usr)
			// TODO: Figure when & where to add & remove users to the user manager
			// TODO: message for user ws: DISCONNECTED (eg. if they go onto another page) (TODO: Use the close ctrl message)
		})

		r.Get("/games/{gameID}/ws", func(w http.ResponseWriter, r *http.Request) {
			// Upgrade our raw HTTP connection to a websocket based one
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			usr := r.Context().Value(CTX_USR).(*User)
			if usr.Conn != nil {
				// TODO: Write reason
				usr.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Time{})
			}
			usr.Conn = conn
			UserMgr.Add(usr)
		})
	})

	return r
}
