package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Node struct {
	NodeId string
}

func NewNode(id string) Node {
	return Node{
		NodeId: id,
	}
}

func (n *Node) InitOk() {
	msg := Message{
		Src: n.NodeId,
		Body: MessageBody{
			Type: "init_ok",
			// InReplyTo: ,
		},
	}
}

// set up a messages package?
type Message struct {
	Src  string
	Dest string
	Body MessageBody
}

type MessageBody struct {
	Type      string
	MsgId     *uint     `json:"msg_id,omitempty"`
	InReplyTo *uint     `json:"in_reply_to,omitempty"`
	NodeId    *string   `json:"node_id,omitempty"`
	NodeIds   *[]string `json:"node_ids,omitempty"`
}

func main() {
	var node Node
	for {
		var msg Message
		err := json.NewDecoder(os.Stdin).Decode(&msg)
		if err != nil {
			os.Exit(1)
		}

		// just pass the message to the node
		// then you can test
		if msg.Body.Type == "init" {
			node = NewNode(*msg.Body.NodeId)
			node.InitOk()
		}
	}
}
