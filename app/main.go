package main

import (
	"bufio"
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

func (n *Node) Init() {
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

func (n *Node) Echo(msg *Message) {
	reply := Message{
		Src:  n.NodeId,
		Dest: msg.Src,
		Body: MessageBody{
			Type:      "echo_ok",
			InReplyTo: msg.Body.MsgId,
			Echo:      msg.Body.Echo,
		},
	}

	json.NewEncoder(os.Stdout).Encode(reply)
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
	Echo      *string   `json:"echo,omitempty"`
	NodeIds   *[]string `json:"node_ids,omitempty"`
}

func main() {
	var node Node
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		jsonString := scanner.Text()
		go func(jsonString string) {
			var msg Message
			err := json.Unmarshal([]byte(jsonString), &msg)
			if err != nil {
				os.Exit(1)
			}

			// pass a reader to the node from which to read the messages
			// then you can test
			// or maybe take bot an input and an output
			if msg.Body.Type == "init" {
				node = NewNode(&msg)
				node.Init()
			}

			if msg.Body.Type == "echo" {
				node.Echo(&msg)
			}
		}(jsonString)
	}
}
