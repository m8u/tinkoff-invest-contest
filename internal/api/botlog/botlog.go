package botlog

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"tinkoff-invest-contest/internal/utils"
)

var prefixRegex = regexp.MustCompile(`\[bot#[0-9]+\]`)
var botIdRegex = regexp.MustCompile(`[0-9]+`)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = struct {
	mu    sync.Mutex
	table map[string]*websocket.Conn
}{
	table: make(map[string]*websocket.Conn),
}

type writer struct{}

var Writer *writer
var logArchiveFile *os.File

func init() {
	var err error
	logArchiveFile, err = os.OpenFile("log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	utils.MaybeCrash(err)
	err = logArchiveFile.Truncate(0)
	utils.MaybeCrash(err)
}

func Echo(c *gin.Context) {
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	defer conn.Close()

	botId := c.Query("id")
	clients.mu.Lock()
	clients.table[botId] = conn
	clients.mu.Unlock()

	writeArchive(botId)

	for {
		mt, _, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			break
		}
	}
}

func (writer *writer) Write(p []byte) (n int, err error) {
	_, err = logArchiveFile.Write(p)
	if err != nil {
		return 0, err
	}

	prefix := string(prefixRegex.Find(p))
	if prefix == "" {
		return len(p), nil
	}
	conn, ok := clients.table[botIdRegex.FindString(prefix)]
	if ok {
		err := conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func writeArchive(botId string) {
	file, err := os.Open("log")
	if err != nil {
		log.Println("error: can't open log archive file")
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		prefix := prefixRegex.FindString(s)
		if prefix != "" && botIdRegex.FindString(prefix) == botId {
			err := clients.table[botId].WriteMessage(websocket.TextMessage, []byte(s))
			if err != nil {
				log.Println("error: failed to write log archive for bot#" + botId)
			}
		}
	}
}
