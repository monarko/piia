package helpers

import (
    "bytes"
    "context"
    "encoding/gob"
    "fmt"
    "log"
    "sync"

    "cloud.google.com/go/pubsub"
)

// GetBytes returns a byte array from an interface
func GetBytes(key interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(key)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// PubSubPublish function
func PubSubPublish(projectID, topicID string, msg interface{}) error {
    msgB, err := GetBytes(msg)
    if err != nil {
        return fmt.Errorf("interface to byte conversion: %v", err)
    }

    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        return fmt.Errorf("pubsub.NewClient: %v", err)
    }

    t := client.Topic(topicID)
    result := t.Publish(ctx, &pubsub.Message{
        Data: msgB,
    })
    // Block until the result is returned and a server-generated
    // ID is returned for the published message.
    id, err := result.Get(ctx)
    if err != nil {
        return fmt.Errorf("pubsub Get: %v", err)
    }
    log.Println("Published a message; msg ID: ", id)
    return nil
}

// PubSubPullMessages function
func PubSubPullMessages(projectID, subID string) error {
    // projectID := "my-project-id"
    // subID := "my-sub"
    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        return fmt.Errorf("pubsub.NewClient: %v", err)
    }

    // Consume 10 messages.
    var mu sync.Mutex
    received := 0
    sub := client.Subscription(subID)
    cctx, cancel := context.WithCancel(ctx)
    err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
        // TODO: process msg
        msg.Ack()
        mu.Lock()
        defer mu.Unlock()
        received++
        if received == 10 {
            cancel()
        }
    })
    if err != nil {
        return fmt.Errorf("Receive: %v", err)
    }
    return nil
}
