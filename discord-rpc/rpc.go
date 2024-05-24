package discordrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/google/uuid"
)

// Discord RPC module
// IPC code taken from https://github.com/rikkuness/discord-rpc

type Command struct {
	Nonce   string         `json:"nonce"`
	Args    map[string]any `json:"args"`
	Command string         `json:"cmd"`
}

type ActivityEmoji struct {
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

type ActivityParty struct {
	ID   string  `json:"id,omitempty"`
	Size *[2]int `json:"size,omitempty"`
}

type ActivityAssets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type ActivitySecrets struct {
	Join     string `json:"join,omitempty"`
	Spectate string `json:"spectate,omitempty"`
	Match    string `json:"match,omitempty"`
}

type ActivityButtons struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type ActivityTimestamps struct {
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

type ActivityType int

const (
	Game ActivityType = iota
	Streaming
	Listening
	Watching
	Custom
	Competiting
)

type ActivityFlags int

const (
	Instance ActivityFlags = 1 << iota
	Join
	Spectate
	JoinRequest
	Sync
	Play
	PartyPrivacyFriends
	PartyPrivacyVoiceChannel
	Embedded
)

type Activity struct {
	//Name          string             `json:"name"`
	Type ActivityType `json:"type"`
	URL  string       `json:"url,omitempty"`
	//CreatedAt     int64              `json:"created_at"`
	Timestamps    ActivityTimestamps `json:"timestamps,omitempty"`
	ApplicationID string             `json:"application_id,omitempty"`
	Details       string             `json:"details,omitempty"`
	State         string             `json:"state,omitempty"`
	//Emoji ActivityEmoji `json:"emoji,omitempty"`
	Party    ActivityParty     `json:"party,omitempty"`
	Assets   ActivityAssets    `json:"assets,omitempty"`
	Secrets  ActivitySecrets   `json:"secrets,omitempty"`
	Instance bool              `json:"instance,omitempty"`
	Flags    ActivityFlags     `json:"flags,omitempty"`
	Buttons  []ActivityButtons `json:"buttons,omitempty"`
}

func (c Command) json() string {
	d, _ := json.Marshal(c)
	return string(d)
}

type Client struct {
	conn     net.Conn
	ClientID string
}

func (c *Client) handshake() (string, error) {
	return c.Send(0, fmt.Sprintf(`{"v":"1", "client_id":"%s"}`, c.ClientID))
}

func (c *Client) sendCommand(op int, command Command) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}
	return c.Send(op, command.json())
}

func (c *Client) Connect() error {
	conn, err := newConnection()
	if err != nil {
		return err
	}
	c.conn = conn
	_, err = c.handshake()
	return err
}

func (c *Client) SetActivity(a Activity) (string, error) {
	return c.sendCommand(1, Command{
		Command: "SET_ACTIVITY",
		Nonce:   uuid.New().String(),
		Args: map[string]any{
			"pid":      os.Getpid(),
			"activity": a,
		},
	})
}

func (c *Client) Read() (string, error) {
	buf := make([]byte, 512)
	payloadlength, err := c.conn.Read(buf)
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	payload := buf[8:payloadlength]
	if _, err := buffer.Write(payload); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (c *Client) Write(op int, data string) error {
	var buf = new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint32(op)); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}

	if _, err := buf.WriteString(data); err != nil {
		return err
	}
	_, err := c.conn.Write(buf.Bytes())
	return err
}

func (c *Client) Send(op int, data string) (string, error) {
	if err := c.Write(op, data); err != nil {
		return "", err
	}
	return c.Read()
}
