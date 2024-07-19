package gorilla

import (
	"loveair/base/cache"
	"loveair/base/data"
	"loveair/base/meta"
	"loveair/email"
	"loveair/log"
	"loveair/models"
	"loveair/push"
	"net/http"
	"sort"
	"time"

	"loveair/core/websocket/router"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/xid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1
)

type Socket struct {
	// The number of connected clients, max should be 200k. Once it passed 200k drop all incoming  Conn till someone leaves.
	connCount int
	clients   cmap.ConcurrentMap[string, *models.Client]
	// join is a channel for admin online.
	join chan *models.Client
	// leave is a channel for clients going ofline.
	leave   chan *models.Client
	dbase   data.Interface
	mbase   meta.Interface
	cbaseIf cache.Interface
	sRouter *router.Router
	emailIf email.Interface
	pushIf  push.Interface
	sLogger log.SLoger
}

func InitWebsocket(dbase data.Interface, mbase meta.Interface, cbaseIf cache.Interface, sRouter *router.Router, emailIf email.Interface, pushIf push.Interface, serviceLogger log.SLoger) *Socket {

	return &Socket{
		connCount: 0,
		clients:   cmap.New[*models.Client](),
		join:      make(chan *models.Client),
		leave:     make(chan *models.Client),
		dbase:     dbase,
		mbase:     mbase,
		cbaseIf:   cbaseIf,
		sRouter:   sRouter,
		emailIf:   emailIf,
		pushIf:    pushIf,
		sLogger:   serviceLogger,
	}
}

func (s *Socket) generateID() string {
	// Generate unique ID.
	uid := xid.New()
	return uid.String()
}

func (s *Socket) Daemon() {
	//~keeps logging the number of connected users
	go func() {
		// Create a new ticker that ticks every 30 seconds
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop() // Ensure the ticker is stopped when no longer needed

		for range ticker.C {
			s.sLogger.Log.Infoln("Clients connected: ", s.clients.Count())
		}
	}()

	// Asyncynchronouse edit of concurrent map (Set & Remove)
	go func() {
		for {
			client := <-s.join

			go func(client *models.Client) {

				// Add client to local map
				s.clients.Set(client.ID, client)

				//add user presence in meta database as online.
				err := s.mbase.UpdateUserPresence(client.ID, "Online", time.Time{})
				if err != nil {
					s.sLogger.Log.Errorln(err)
				}

				// Add client information to cache, so that other instance will know they are connected to this instance.
				// Step 1: Check if client exist
				if ok, err := s.cbaseIf.ClientExist(client.ID); err == nil && ok {
					// Step 2: if cache returns no error and client existence in cache is positive
					//Update the cache to reflect the clients new instance.
					// if err := s.cbaseIf.UpdateClientInstanceID(client.ID, ch.router.GetInstanceID()); err != nil {
					// 		ctxLogger.Errorln(err)
					// }
					// Step 3: if cache returns no error but the client existence is negative
					// Add the client to cache.
				} else if err == nil && !ok {
					if err := s.cbaseIf.AddClient(client.ID, models.ClientCache{InstanceID: "0", CachedChat: []string{}}); err != nil {
						s.sLogger.Log.Errorln(err)
					}
				} else {
					// Step 4: in the case of a cache error, log it.
					s.sLogger.Log.Errorln(err)
				}
			}(client)
		}
	}()

	go func() {
		for {
			client := <-s.leave
			go func(client *models.Client) {
				// Remove client from local map.
				s.clients.Remove(client.ID)

				//add user presence in meta database as offline.
				err := s.mbase.UpdateUserPresence(client.ID, "Offline", time.Now().UTC())
				if err != nil {
					s.sLogger.Log.Errorln(err)
				}

				// Remove client information from cache & cache open session,
				if ccIDs, err := s.cbaseIf.GetClientCachedChatSlice(client.ID); err == nil {
					go s.persisteCachedChats(*ccIDs)

					if err := s.cbaseIf.RemoveClient(client.ID); err != nil {
						s.sLogger.Log.Errorln(err)
					}
				} else {
					//!handle error properly
					s.sLogger.Log.Errorln(err)
				}

			}(client)
		}
	}()
}

// persisteCachedChats persist all user open chats when user goes offline or changes instance.
func (s *Socket) persisteCachedChats(cIDs []string) {
	for _, id := range cIDs {
		if cs, err := s.cbaseIf.GetCachedChat(id); err == nil {
			go s.persiste(cs.Messages)
			s.cbaseIf.DeleteCachedChat(cs.Messages[0].ChatID)
		} else {
			s.sLogger.Log.Errorln(err)

		}
	}
}

func (s *Socket) persiste(msgs []models.Message) {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[j].Timestamp.Before(msgs[i].Timestamp)
	})

	err := s.dbase.MergeCachedSession(msgs)
	if err != nil {
		//!handle error properly
		s.sLogger.Log.Errorln(err)
		return
	}
}
