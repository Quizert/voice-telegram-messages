package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "kursach/kursach/proto" // Замени на свой путь

	"google.golang.org/grpc"
)

func SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error) {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		fmt.Printf("did not connect: %v", err)
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewAudioProcessorClient(conn)

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
