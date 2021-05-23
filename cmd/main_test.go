package main

import (
	"context"
	"encoding/json"
	"postgres_pub_sub/postgres"
	"testing"
)

type Payload struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type notification struct {
	Op      string  `json:"op,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

func Test_run(t *testing.T) {
	type args struct {
		wantEvent notification
		operation string
	}

	postgresClient := postgres.NewTestClient(t)

	defer func() {
		postgresClient.Close()
		postgresClient.DeleteTestDatabase(t)
	}()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert and catch event",
			args: args{
				wantEvent: notification{
					Op:      "INSERT",
					Payload: Payload{ID: 1, Name: "event_1"},
				},
				operation: "insert",
			},
			wantErr: false,
		},
		{
			name: "update and catch event",
			args: args{
				wantEvent: notification{
					Op: "UPDATE",
					Payload: Payload{
						ID:   1,
						Name: "event_1_edited",
					},
				},
				operation: "update",
			},
			wantErr: false,
		},
		{
			name: "delete and catch event",
			args: args{
				wantEvent: notification{
					Op: "DELETE",
					Payload: Payload{
						ID:   1,
						Name: "event_1_edited",
					},
				},
				operation: "delete",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notchan, err := postgresClient.ListenToEvents(context.TODO(), "event")
			if err != nil {
				t.Fatal("failed to listen to events with: ", err.Error())
			}
			switch tt.args.operation {
			case "insert":
				err = postgresClient.AddEvent(context.TODO(), "event_1")
				if err != nil {
					t.Fatal(err.Error())
				}
			case "update":
				err = postgresClient.UpdateEventByID(context.TODO(), "event_1_edited", 1)
				if err != nil {
					t.Fatal(err.Error())
				}
			case "delete":
				err = postgresClient.DeleteEventByID(context.TODO(), 1)
				if err != nil {
					t.Fatal(err.Error())
				}
			}
			not := <-notchan

			var event notification
			err = json.Unmarshal([]byte(not.Payload), &event)
			if err != nil {
				t.Error(err.Error())
			}
			t.Logf("got notification: %+v", event)
		})
	}
}
