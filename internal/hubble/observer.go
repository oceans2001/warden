package hubble

import (
	"context"
	"io"
	"log"

	observer "github.com/cilium/cilium/api/v1/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FlowEvent struct {
	Session     string
	Verdict     string
	Destination string
}

func StartStream(store *Store) {
	conn, err := grpc.NewClient("localhost:4245", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Failed to connect to Hubble Relay:", err)
		return
	}

	client := observer.NewObserverClient(conn)

	stream, err := client.GetFlows(context.Background(), &observer.GetFlowsRequest{
		Follow: true,
	})
	if err != nil {
		log.Println("Failed to start flow stream:", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("Hubble stream closed")
			return
		}
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		flow := resp.GetFlow()
		if flow == nil {
			continue
		}

		session := extractSessionLabel(flow.GetSource().GetLabels())
		if session == "" {
			continue
		}

		destination := ""
		if names := flow.GetDestinationNames(); len(names) > 0 {
			destination = names[0]
		}

		store.Add(FlowEvent{
			Session:     session,
			Verdict:     flow.GetVerdict().String(),
			Destination: destination,
		})
	}
}

func extractSessionLabel(labels []string) string {
	prefix := "k8s:warden-session="
	for _, l := range labels {
		if len(l) > len(prefix) && l[:len(prefix)] == prefix {
			return l[len(prefix):]
		}
	}
	return ""
}
