package ovs

import "github.com/docker/docker/client"

type dockerer struct {
	client *client.Client
}
