package service

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

func StreamEvents(cli *client.Client, ctx context.Context, eventsCh chan events.Message, errsCh chan error) {
	msgs, errs := cli.Events(ctx, events.ListOptions{})
	for {
		select {
		case e := <-msgs:
			eventsCh <- e
		case err := <-errs:
			errsCh <- err
		case <-ctx.Done():
			return
		}
	}
}
