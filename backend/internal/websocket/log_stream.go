package websocket

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsWriter struct {
	conn *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	if err := w.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func StreamLogs(cli *client.Client, w http.ResponseWriter, r *http.Request, containerID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	ctx := context.Background()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	}

	reader, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	defer reader.Close()

	writer := &wsWriter{conn: conn}

	// Stream logs continuously
	if _, err := stdcopy.StdCopy(writer, writer, reader); err != nil && err != io.EOF {
		// only log if it's a real error, ignore client disconnects
		fmt.Println("log stream error:", err)
	}
}
