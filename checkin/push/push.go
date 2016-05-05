package mdmpush

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/push"
	"github.com/go-kit/kit/log"
)

// Pusher pushes things
type Pusher interface {
	Push(magic, token string)
}

//New returns a Pusher
func New(logger log.Logger, certPath, password string) Pusher {
	filename := certPath
	cert, key, err := certificate.Load(filename, password)
	if err != nil {
		fmt.Println(err)
		logger.Log("err", err)
	}
	client, err := push.NewClient(certificate.TLS(cert, key))
	if err != nil {
		fmt.Println(err)
		logger.Log("err", err)
	}
	service := &push.Service{
		Client: client,
		Host:   push.Production,
	}

	pusher := pushSvc{
		client: service,
		logger: logger,
	}

	return pusher
}

type pushSvc struct {
	client *push.Service
	logger log.Logger
}

func (svc pushSvc) Push(magic, token string) {
	svc.logger.Log("action", "push", "magic", magic, "token", token)
	enc := token

	p := payload.MDM{Token: magic}
	json.NewEncoder(os.Stdout).Encode(p)
	valid := push.IsDeviceTokenValid(enc)
	if !valid {
		svc.logger.Log("err", "Invalid Token")
		return
	}
	id, err := svc.client.Push(enc, nil, p)
	if err != nil {
		svc.logger.Log("err", err)
		return
	}
	svc.logger.Log("action", "push", "magic", magic, "token", token, "id", id)
	fmt.Println("success")
}
