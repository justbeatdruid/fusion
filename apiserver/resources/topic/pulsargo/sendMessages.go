package pulsargo

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
)
func SendMessages(client pulsar.Client,topicUrl string,messagesBody string,key string) (pulsar.MessageID,error) {
	defer client.Close()
    produce,err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   topicUrl,
		Name:                    "",
	})
    defer produce.Close()
	if err!=nil {
		return nil,fmt.Errorf("create pulsar producer error: %s",err)
	}
    messageId,err := produce.Send(context.Background(),&pulsar.ProducerMessage{
		Payload:             []byte(messagesBody),
		Key:                 key,
	})
	if err!=nil {
		return nil,fmt.Errorf("send messages error: %s",err)
	}
	id,_:= pulsar.DeserializeMessageID(messageId.Serialize())
    fmt.Println(id)
return messageId,nil

}
