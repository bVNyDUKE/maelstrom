package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Message struct {
	Src  string      `json:"src"`
	Dest string      `json:"dest,omitempty"`
	Body MessageBody `json:"body"`
}

type MessageBody struct {
	Type      string               `json:"type"`
	MsgId     *uint                `json:"msg_id,omitempty"`
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
	NodeId    string
	LastMsg   *Message
	Stdin     io.Reader
	Stdout    io.Writer
	resChan   chan Message
	messages  []int
	neighbors []string
}

func NewNode() *Node {
	return &Node{
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		resChan:   make(chan Message),
		messages:  []int{},
		neighbors: []string{},
	}
}

func (n *Node) res(msg *Message, body *MessageBody) error {
	n.LastMsg = msg
	body.InReplyTo = n.LastMsg.Body.MsgId

	res := Message{
		Src:  n.NodeId,
		Dest: n.LastMsg.Src,
		Body: *body,
	}

	n.resChan <- res
	return nil
}

func (n *Node) send(dest string, body *MessageBody) {
	res := Message{
		Src:  n.NodeId,
		Dest: dest,
		Body: *body,
	}

	n.resChan <- res
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
	id := fmt.Sprintf("id-%s-%d", n.NodeId, *msg.Body.MsgId)
	n.res(msg, &MessageBody{
		Type: "generate_ok",
		Id:   &id,
	})
}

func (n *Node) broadcast(msg *Message) {
	content := *msg.Body.Message

	for c := range n.messages {
		if c == content {
			n.res(msg, &MessageBody{
				Type: "broadcast_ok",
			})
			return
		}
	}

	n.messages = append(n.messages, content)

	for _, dest := range n.neighbors {
		n.send(dest, &MessageBody{
			Type:    "broadcast",
			Message: &content,
		})
	}

	n.res(msg, &MessageBody{
		Type: "broadcast_ok",
	})
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
		InReplyTo: msg.Body.MsgId,
	})
}

func (n *Node) handleResponses() {
	for {
		select {
		case msg := <-n.resChan:
			json.NewEncoder(n.Stdout).Encode(msg)
		}
	}
}

// i guess i need a channel to receive the messages
// and also one more channel to get the responses
// that way i can avoid a mutex?
func (n *Node) Run() error {
	go n.handleResponses()

	scanner := bufio.NewScanner(n.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			os.Exit(1)
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
