package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"slices"
	"sync"
	"time"
	// "time"
)

type Message struct {
	Src  string      `json:"src"`
	Dest string      `json:"dest,omitempty"`
	Body MessageBody `json:"body"`
}

type MessageBody struct {
	Type      string               `json:"type"`
	MsgId     uint                 `json:"msg_id,omitempty"`
	InReplyTo *uint                `json:"in_reply_to,omitempty"`
	NodeId    *string              `json:"node_id,omitempty"`
	Echo      *string              `json:"echo,omitempty"`
	NodeIds   *[]string            `json:"node_ids,omitempty"`
	Id        *string              `json:"id,omitempty"`
	Message   *int                 `json:"message,omitempty"`
	Messages  *[]int               `json:"messages,omitempty"`
	Topology  *map[string][]string `json:"topology,omitempty"`
}

type Node struct {
	NodeId string
	Stdin  io.Reader
	Stdout io.Writer

	wg        sync.WaitGroup
	mut       sync.Mutex
	nextMsgId uint
	messages  []int
	neighbors []string
	nodeIds   []string
	callbacks map[uint]func(*Message)
}

func NewNode() *Node {
	return &Node{
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		nextMsgId: 0,
		messages:  []int{},
		neighbors: []string{},
		callbacks: make(map[uint]func(*Message)),
	}
}

func (n *Node) addStoredMessages(newMessage ...int) {
	n.mut.Lock()
	n.messages = append(n.messages, newMessage...)
	n.mut.Unlock()
}

func (n *Node) popCallbackHandler(key uint) func(*Message) {
	n.mut.Lock()
	handler := n.callbacks[key]
	delete(n.callbacks, key)
	n.mut.Unlock()

	return handler
}

func (n *Node) handleRes(msg *Message) {
	log.Println("Sending message in reply to", msg.Body.Type)

	n.mut.Lock()
	msg.Body.MsgId = n.nextMsgId
	n.nextMsgId++
	n.mut.Unlock()

	json.NewEncoder(n.Stdout).Encode(msg)
}

func (n *Node) res(msg *Message, body *MessageBody) {
	body.InReplyTo = &msg.Body.MsgId

	res := Message{
		Src:  n.NodeId,
		Dest: msg.Src,
		Body: *body,
	}

	n.handleRes(&res)
}

func (n *Node) send(dest string, body *MessageBody) {
	res := Message{
		Src:  n.NodeId,
		Dest: dest,
		Body: *body,
	}

	n.handleRes(&res)
}

func (n *Node) init(msg *Message) {
	n.NodeId = *msg.Body.NodeId

	n.mut.Lock()
	n.nodeIds = *msg.Body.NodeIds
	n.mut.Unlock()

	n.res(msg, &MessageBody{
		Type: "init_ok",
	})
	log.Println("Node initialized")
}

func (n *Node) echo(msg *Message) {
	n.res(msg, &MessageBody{
		Type: "echo_ok",
		Echo: msg.Body.Echo,
	})
}

func (n *Node) generate(msg *Message) {
	id := fmt.Sprintf("id-%s-%d", n.NodeId, msg.Body.MsgId)
	n.res(msg, &MessageBody{
		Type: "generate_ok",
		Id:   &id,
	})
}

func (n *Node) broadcast(msg *Message) {
	n.res(msg, &MessageBody{
		Type: "broadcast_ok",
	})
	content := *msg.Body.Message
	if slices.Contains(n.messages, content) {
		return
	}

	n.addStoredMessages(content)
}

func (n *Node) read(msg *Message) {
	n.res(msg, &MessageBody{
		Type:     "read_ok",
		Messages: &n.messages,
	})
}

func (n *Node) topology(msg *Message) {
	topology := *msg.Body.Topology
	if topology == nil {
		log.Fatal("No topology in message")
	}
	neighbors, ok := topology[n.NodeId]
	if !ok {
		log.Fatalf("No neighbors for node: %s", n.NodeId)
	}
	n.neighbors = neighbors

	n.res(msg, &MessageBody{
		Type:      "topology_ok",
		InReplyTo: &msg.Body.MsgId,
	})
}

func (n *Node) parseGossip(msg *Message) {
	newMsgs := make([]int, 0, 20)
	for _, msg := range *msg.Body.Messages {
		if slices.Contains(n.messages, msg) {
			continue
		}
		newMsgs = append(newMsgs, msg)
	}
	n.addStoredMessages(newMsgs...)
}

func (n *Node) sendGossip() {
	if len(n.messages) < 2 {
		return
	}

	for _, node := range n.nodeIds {
		if rand.Intn(2) != 1 {
			continue
		}

		n.send(node, &MessageBody{
			Type:     "gossip",
			Messages: &n.messages,
		})
	}
}

func (n *Node) startGossiping(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n.sendGossip()
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
	}
}

func (n *Node) Run() error {
	scanner := bufio.NewScanner(n.Stdin)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go n.startGossiping(ctx)

	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Fatalf("Error deserializing message, %T", line)
		}

		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			if msg.Body.Type == "init" {
				n.init(&msg)
			}
			if msg.Body.Type == "echo" {
				n.echo(&msg)
			}
			if msg.Body.Type == "generate" {
				n.generate(&msg)
			}
			if msg.Body.Type == "broadcast" {
				n.broadcast(&msg)
			}
			if msg.Body.Type == "read" {
				n.read(&msg)
			}
			if msg.Body.Type == "topology" {
				n.topology(&msg)
			}
			if msg.Body.Type == "gossip" {
				n.parseGossip(&msg)
			}
			return
		}()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	n.wg.Wait()

	return nil
}
