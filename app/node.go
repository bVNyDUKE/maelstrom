package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

type Node struct {
	NodeId  string
	LastMsg *Message
	Stdin   io.Reader
}

func NewNode() *Node {
	return &Node{
		Stdin: os.Stdin,
	}
}

func (n *Node) Init(msg *Message) {
	n.NodeId = *msg.Body.NodeId
	n.LastMsg = msg

	res := Message{
		Src:  n.NodeId,
		Dest: n.LastMsg.Src,
		Body: MessageBody{
			Type:      "init_ok",
			InReplyTo: n.LastMsg.Body.MsgId,
		},
	}

	json.NewEncoder(os.Stdout).Encode(res)
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

func (n *Node) Run() error {
	scanner := bufio.NewScanner(n.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			os.Exit(1)
		}
		go func() {
			// pass a reader to the node from which to read the messages
			// then you can test
			// or maybe take bot an input and an output
			if msg.Body.Type == "init" {
				n.Init(&msg)
			}

			if msg.Body.Type == "echo" {
				n.Echo(&msg)
			}
		}()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

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
