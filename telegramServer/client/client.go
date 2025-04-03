package client

import (
	"context"
	"fmt"
	"time"

	pb "kursach/proto" // Замени на свой путь

	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
}

type AudioProcessorClient struct {
	conn *grpc.ClientConn
}

func NewAudioProcessorClient(host, port string) (*AudioProcessorClient, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
	}

	return &AudioProcessorClient{conn: conn}, nil
}

func (a *AudioProcessorClient) SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error) {
	client := pb.NewAudioProcessorClient(a.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	req := &pb.ContentRequest{
		Text: text,
		Audio: &pb.AudioFile{
			Data: audioData,
		},
	}

	resp, err := client.ProcessContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке: %v", err)
	}

	return resp, nil
}
