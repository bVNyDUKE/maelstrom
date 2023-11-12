package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
	Id        *string   `json:"id,omitempty"`
}

type Node struct {
	NodeId  string
	LastMsg *Message
	Stdin   io.Reader
	Stdout  io.Writer
}

func NewNode() *Node {
	return &Node{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}
}

func (n *Node) init(msg *Message) {
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

	json.NewEncoder(n.Stdout).Encode(res)
}

func (n *Node) echo(msg *Message) {
	n.LastMsg = msg

	reply := Message{
		Src:  n.NodeId,
		Dest: msg.Src,
		Body: MessageBody{
			Type:      "echo_ok",
			InReplyTo: msg.Body.MsgId,
			Echo:      msg.Body.Echo,
		},
	}

	json.NewEncoder(n.Stdout).Encode(reply)
}

func (n *Node) generate(msg *Message) {
	n.LastMsg = msg
	id := fmt.Sprintf("id-%s-%d", n.NodeId, *msg.Body.MsgId)

	reply := Message{
		Src:  n.NodeId,
		Dest: msg.Src,
		Body: MessageBody{
			Type:      "generate_ok",
			InReplyTo: msg.Body.MsgId,
			Id:        &id,
		},
	}

	json.NewEncoder(n.Stdout).Encode(reply)
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
			if msg.Body.Type == "init" {
				n.init(&msg)
			}

			if msg.Body.Type == "echo" {
				n.echo(&msg)
			}
			if msg.Body.Type == "generate" {
				n.generate(&msg)
			}
		}()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
