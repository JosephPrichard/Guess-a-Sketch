/*
 * Copyright (c) Joseph Prichard 2024
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"guessthesketch/game"
	"guessthesketch/servers"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const WsHost = "ws://localhost:8080"
const HttpHost = "http://localhost:8080"

func createRoom(playerCount int) servers.RoomCodeResp {
	url := fmt.Sprintf("%s/api/rooms/create", HttpHost)

	roomSettings := game.RoomSettings{
		PlayerLimit:   playerCount,
		TimeLimitSecs: 15,
		TotalRounds:   1,
	}
	body, err := servers.PostJson(url, roomSettings)
	if err != nil {
		log.Fatalf("Failed to post json %v", err)
	}

	var roomResp servers.RoomCodeResp
	err = json.Unmarshal(body, &roomResp)
	if err != nil {
		log.Fatalf("Failed to unmarshal json %v", err)
	}
	return roomResp
}

func joinPlayersToRoom(code string, count int) []*websocket.Conn {
	// create connections and join a room for each player client
	websockets := make([]*websocket.Conn, 0)
	for i := 0; i < count; i++ {
		u := fmt.Sprintf("%s/api/rooms/join?code=%s", WsHost, code)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err != nil {
			log.Fatalf("%v", err)
		}
		websockets = append(websockets, ws)
	}
	return websockets
}

func sendStartMessage(ws *websocket.Conn) {
	input := game.InputPayload[struct{}]{
		Code: game.StartCode,
	}
	buf, err := json.Marshal(input)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		log.Printf("Failed to send start message %v", err)
		return
	}
}

func sendDrawMessage(ws *websocket.Conn, traceID string) {
	input := game.InputPayload[game.DrawMsg]{
		Code: game.DrawCode,
		Msg: game.DrawMsg{
			X:      uint16(rand.Intn(game.MaxX)),
			Y:      uint16(rand.Intn(game.MaxY)),
			Radius: 1,
			Color:  1,
		},
		TraceID: traceID,
	}
	buf, err := json.Marshal(input)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		log.Printf("Failed to send start message %v", err)
		return
	}
}

type LogMsg struct {
	key  string
	time int
}

type BenchmarkCtx struct {
	mu          sync.Mutex
	messagesIn  []LogMsg
	messagesOut []LogMsg
}

func (ctx *BenchmarkCtx) addMsgIn(m LogMsg) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.messagesIn = append(ctx.messagesIn, m)
}

func (ctx *BenchmarkCtx) addMsgOut(m LogMsg) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.messagesOut = append(ctx.messagesOut, m)
}

func (ctx *BenchmarkCtx) runRoomClient(playerCount int, roomsWg *sync.WaitGroup) {
	defer roomsWg.Done()

	if playerCount < 1 {
		log.Fatalf("Cannot start a room with less than 1 player")
	}

	roomResp := createRoom(playerCount)
	websockets := joinPlayersToRoom(roomResp.Code, playerCount)

	// wait on all player clients to finish
	var playersWg sync.WaitGroup
	playersWg.Add(playerCount)

	sendStartMessage(websockets[0])
	for i := 0; i < playerCount; i++ {
		clientKey := fmt.Sprintf("%s_%d", roomResp.Code, i)
		go ctx.runPlayerClient(websockets[i], clientKey, &playersWg)
	}

	playersWg.Wait()
}

func (ctx *BenchmarkCtx) runPlayerClient(ws *websocket.Conn, clientKey string, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Second)

	go func(ws *websocket.Conn) {
		defer func() {
			wg.Done()
			ticker.Stop()
		}()
		for {
			_, buf, err := ws.ReadMessage()
			if err != nil {
				return
			}

			var payload game.OutputPayload[json.RawMessage]
			err = json.Unmarshal(buf, &payload)
			if err != nil {
				log.Fatalf("Failed to unmarshal output payload")
			}

			if payload.Code == game.FinishCode {
				var msg game.FinishMsg
				err := json.Unmarshal(payload.Msg, &msg)
				if err != nil {
					log.Fatalf("Failed to unmarshall msg for finish payload")
				}
				if msg.BeginMsg == nil {
					// this means there is no next "next" round, so the client can stop listening
					return
				}
			}

			if payload.TraceID != "" {
				msg := LogMsg{
					key:  payload.TraceID,
					time: int(time.Now().UnixMicro()),
				}
				ctx.addMsgOut(msg)

				log.Printf("Received a message %s with key %s", string(buf), msg.key)
			}
		}
	}(ws)

	trace := 0
	for range ticker.C {
		traceID := clientKey + "_" + strconv.Itoa(trace)
		msg := LogMsg{
			key:  traceID,
			time: int(time.Now().UnixMicro()),
		}
		ctx.addMsgIn(msg)

		sendDrawMessage(ws, traceID)

		log.Printf("Player client %s sent a message with trace %s", clientKey, traceID)
		trace += 1
	}

	wg.Done()
}

// aggregates messages by key
func aggregateMessages(messages []LogMsg) map[string][]LogMsg {
	aggMsg := make(map[string][]LogMsg)
	for _, m := range messages {
		if aggMsg[m.key] != nil {
			aggMsg[m.key] = append(aggMsg[m.key], m)
		} else {
			aggMsg[m.key] = []LogMsg{m}
		}
	}
	return aggMsg
}

// calculate all latencies that log them out, the return the average latency
func calcLatencies(aggMsgOut map[string][]LogMsg, msgLogIn []LogMsg) int {
	totalLatency := 0
	count := 0

	for _, mIn := range msgLogIn {
		msgLogOut, ok := aggMsgOut[mIn.key]
		if !ok {
			continue
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Latencies for message %s:", mIn.key))
		for _, mOut := range msgLogOut {
			latency := mOut.time - mIn.time
			totalLatency += latency
			count += 1

			sb.WriteString(fmt.Sprintf(" %dms ", latency/1e3))
		}
		fmt.Println(sb.String())
	}

	if count == 0 {
		return 0
	}
	return totalLatency / count
}

func logThroughput(msgLog []LogMsg) {
	startTime := msgLog[0].time
	endTime := msgLog[len(msgLog)-1].time
	duration := float64(endTime-startTime) / 1e6
	throughput := float64(len(msgLog)) / duration

	fmt.Printf("Duration: %f secs\nMessages: %d msgs\nThroughput: %f msgs/sec\n",
		duration, len(msgLog), throughput)
}

func main() {
	roomCount := 1
	var roomsWg sync.WaitGroup
	var ctx BenchmarkCtx
	ctx.messagesIn = make([]LogMsg, 0)

	for i := 0; i < roomCount; i++ {
		roomsWg.Add(1)
		go ctx.runRoomClient(2, &roomsWg)
	}
	roomsWg.Wait()

	fmt.Println("Finished ctx with the following results:")
	if len(ctx.messagesIn) < 1 || len(ctx.messagesOut) < 1 {
		return
	}

	aggMsgOut := aggregateMessages(ctx.messagesOut)
	avgLatency := calcLatencies(aggMsgOut, ctx.messagesIn)

	fmt.Print("\nMessages In:\n")
	logThroughput(ctx.messagesIn)

	fmt.Print("\nMessages Out:\n")
	logThroughput(ctx.messagesOut)
	fmt.Printf("\nAvg Message Pairing Latency: %dms\n", avgLatency/1e3)
}
