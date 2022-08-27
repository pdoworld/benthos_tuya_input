package input

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	pulsar "github.com/tuya/tuya-pulsar-sdk-go"
	"github.com/tuya/tuya-pulsar-sdk-go/pkg/tylog"
	"github.com/tuya/tuya-pulsar-sdk-go/pkg/tyutils"

	"github.com/benthosdev/benthos/v4/public/service"
)

var tuyaConfigSpec = service.NewConfigSpec().
	Summary("Creates an input for tuya pulsar.").
	Field(service.NewStringField("accessId").Default("")).
	Field(service.NewStringField("accessSecert").Default("")).
	Field(service.NewBoolField("debug").Default(false))


func newGibberishInput(conf *service.ParsedConfig) (service.Input, error) {
	accessId, err := conf.FieldString("accessId")
	if err != nil {
		return nil, err
	}
	accessSecert, err := conf.FieldString("accessSecert")
	if err != nil {
		return nil, err
	}
	debug, err := conf.FieldBool("debug")
	if err != nil {
		return nil, err
	}
	if len(accessId) <= 0 {
		return nil, errors.New("accessId is required")
	}
	if len(accessSecert) <= 0 {
		return nil, errors.New("accessSecert is required")

	}
	mes := make(Message)

	tylog.SetGlobalLog("tuya", !debug)

	return &tuyaInput{accessId, accessSecert, debug, "", &mes}, nil
}

func init() {
	err := service.RegisterInput(
		"tuya", tuyaConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			return newGibberishInput(conf)
		})
	if err != nil {
		panic(err)
	}
}

//------------------------------------------------------------------------------
type Message chan string

type tuyaInput struct {
	accessId string
	accessSecert string
	debug bool
	AesSecret string
	mess *Message
}

func (g *tuyaInput) Connect(ctx context.Context) error {
	accessID := g.accessId
	accessKey := g.accessSecert
	topic := pulsar.TopicForAccessID(accessID)

	// create client
	cfg := pulsar.ClientConfig{
		PulsarAddr: pulsar.PulsarAddrCN,
	}
	c := pulsar.NewClient(cfg)

	// create consumer
	csmCfg := pulsar.ConsumerConfig{
		Topic: topic,
		Auth:  pulsar.NewAuthProvider(accessID, accessKey),
	}
	csm, _ := c.NewConsumer(csmCfg)
	g.AesSecret = accessKey[8:24]
	// handle message
	go csm.ReceiveAndHandle(context.Background(), g)
	return nil
}

func (g *tuyaInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	select {
		case slsmsg := <-*g.mess: // read a message from the channel
			msg := service.NewMessage(nil)
			msg.SetStructured(slsmsg)
	
			return msg, func(ctx context.Context, err error) error {
				return nil
			}, nil
		case <-ctx.Done():
			return nil, nil, ctx.Err()
	}
}

func (g *tuyaInput) Close(ctx context.Context) error {
	return nil
}

func (g *tuyaInput) HandlePayload(ctx context.Context, msg *pulsar.Message, payload []byte) error {
	// let's decode the payload with AES
	m := map[string]interface{}{}
	err := json.Unmarshal(payload, &m)
	if err != nil {
		tylog.Error("json unmarshal failed", tylog.ErrorField(err))
		return nil
	}
	bs := m["data"].(string)
	de, err := base64.StdEncoding.DecodeString(string(bs))
	if err != nil {
		tylog.Error("base64 decode failed", tylog.ErrorField(err))
		return nil
	}
	decode := tyutils.EcbDecrypt(de, []byte(g.AesSecret))
	*g.mess <- string(decode)
	return nil
}
