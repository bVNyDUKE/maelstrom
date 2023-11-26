package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
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

	mut       sync.Mutex
	lastMsg   *Message
	nextMsgId uint
	messages  []int
	neighbors []string
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

func (n *Node) handleRes(msg *Message) {
	n.mut.Lock()

	msg.Body.MsgId = n.nextMsgId
	n.nextMsgId++
	json.NewEncoder(n.Stdout).Encode(msg)

	n.mut.Unlock()
}

func (n *Node) res(msg *Message, body *MessageBody) {
	n.mut.Lock()
	n.lastMsg = msg
	n.mut.Unlock()

	body.InReplyTo = &n.lastMsg.Body.MsgId

	res := Message{
		Src:  n.NodeId,
		Dest: n.lastMsg.Src,
		Body: *body,
	}

	n.handleRes(&res)
}

func (n *Node) send(dest string, body *MessageBody, handler *func(*Message)) {
	n.mut.Lock()
	n.callbacks[n.nextMsgId] = *handler
	n.mut.Unlock()

	res := Message{
		Src:  n.NodeId,
		Dest: dest,
		Body: *body,
	}

	n.handleRes(&res)
}

func (n *Node) init(msg *Message) {
	n.NodeId = *msg.Body.NodeId
	n.res(msg, &MessageBody{
		Type: "init_ok",
	})
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

	for _, c := range n.messages {
		if c == content {
			return
		}
	}

	n.messages = append(n.messages, content)

	gossipTo := make([]string, 0, len(n.neighbors))
	for _, neigh := range n.neighbors {
		if neigh != msg.Src {
			gossipTo = append(gossipTo, neigh)
		}
	}

	handler := func(msg *Message) {
		var newGossipTo []string
		for _, des := range gossipTo {
			if des != msg.Src {
				newGossipTo = append(newGossipTo, des)
			}
		}
		gossipTo = newGossipTo
	}

	for len(gossipTo) != 0 {
		for _, dest := range gossipTo {
			n.send(
				dest,
				&MessageBody{
					Type:    "broadcast",
					Message: &content,
				},
				&handler,
			)
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
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

func (n *Node) Run() error {
	scanner := bufio.NewScanner(n.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Fatalf("Error deserializing message, %T", line)
		}

		inReply := msg.Body.InReplyTo

		if inReply != nil {
			n.mut.Lock()
			handler := n.callbacks[*inReply]
			delete(n.callbacks, *inReply)
			n.mut.Unlock()
			if handler != nil {
				go handler(&msg)
			}
			continue
		}

		go func() {
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
		}()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
