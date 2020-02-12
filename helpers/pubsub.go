package helpers

import (
    "bytes"
    "context"
    "encoding/gob"
    "fmt"
    "time"

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
func PubSubPublish(projectID, topicID string, msg interface{}) (string, error) {
    msgB, err := GetBytes(msg)
    if err != nil {
        return "", fmt.Errorf("interface to byte conversion: %v", err)
    }

    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        return "", fmt.Errorf("pubsub.NewClient: %v", err)
    }

    t := client.Topic(topicID)
    result := t.Publish(ctx, &pubsub.Message{
        Data: msgB,
    })
    // Block until the result is returned and a server-generated
    // ID is returned for the published message.
    id, err := result.Get(ctx)
    if err != nil {
        return "", fmt.Errorf("pubsub Get: %v", err)
    }
    return id, nil
}

// PubSubPullMessages function
func PubSubPullMessages(projectID, subID string) ([][]byte, error) {
    // projectID := "my-project-id"
    // subID := "my-sub"
    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        return nil, fmt.Errorf("pubsub.NewClient: %v", err)
    }

    sub := client.Subscription(subID)
    cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    messages := make([][]byte, 0)
    err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
        messages = append(messages, msg.Data)
        msg.Ack()
    })
    if err != nil {
        return nil, fmt.Errorf("receive: %v", err)
    }
    return messages, nil
}
