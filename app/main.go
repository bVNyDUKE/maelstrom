package main

import (
	"encoding/json"
	"os"
)

type Node struct {
	NodeId  string
	LastMsg *Message
}

func NewNode(initMsg *Message) Node {
	return Node{
		NodeId:  *initMsg.Body.NodeId,
		LastMsg: initMsg,
	}
}

func (n *Node) InitOk() {
	msg := Message{
		Src:  n.NodeId,
		Dest: n.LastMsg.Src,
		Body: MessageBody{
			Type:      "init_ok",
			InReplyTo: n.LastMsg.Body.MsgId,
		},
	}

	json.NewEncoder(os.Stdout).Encode(msg)
}

// set up a messages package?
type Message struct {
	Src  string      `json:"src"`
	Dest string      `json:"dest,omitempty"`
	Body MessageBody `json:"body"`
}

type MessageBody struct {
	Type      string    `json:"type"`
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
			node = NewNode(&msg)
			node.InitOk()
		}
	}
}
