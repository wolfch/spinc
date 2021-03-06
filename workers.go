package main

import (
	"fmt"
	"net"
	"time"
)

// Handle webhooks + fetching
func SparkWorker() {
	channels.workers++
	for {
		select {
		case req := <-channels.WhMessage:
			HandleWhMessage(req)
		case req := <-channels.WhMember:
			HandleWhMember(req)
		case req := <-channels.WhRoom:
			HandleWhRoom(req)
		case name := <-channels.Whois:
			WhoisUsers(name)
		case name := <-channels.CreateRoom:
			CreateRoom(name)
		case space_id := <-channels.Members:
			GetMembersOfSpace(space_id)
		case space_id := <-channels.Messages:
			GetMessagesForSpace(space_id)
		case <-channels.Quit:
			return
		}
	}
}

// Update own information
func GetOwnInfo() {
	ticker := time.NewTicker(10 * time.Second)
	channels.workers++
	for {
		select {
		case <-ticker.C:
			GetMeInfo()
		case <-channels.Quit:
			return
		}
	}
}

// Update status time
func UpdateStatusTime() {
	ticker := time.NewTicker(1 * time.Second)
	channels.workers++
	for {
		select {
		case <-ticker.C:
			win.StatusTime.SetText(fmt.Sprintf("[navy][[white]%s[navy]]", time.Now().Format("15:04:05")))
			win.App.Draw()
		case <-channels.Quit:
			return
		}
	}
}

// Check and update latency
func UpdateStatusLag() {
	channels.workers++
	win.StatusLag.SetText(fmt.Sprintf("[navy][[white]Lag: %v[navy]]", "-"))
	count := 0
	for {
		conn, err := net.DialTimeout("tcp", "api.ciscospark.com:80", 5*time.Second)
		if err != nil {
			if (count%120 == 0 || count == 2) && count != 0 {
				AddStatusText("[red]Connection seems to be lost. Retrying every 10 seconds.")
			}
			time.Sleep(1000 * time.Millisecond)
			count++
			continue
		}
		// If success, check if we have been down and if so, perform an update.
		if count > 2 {
			AddStatusText(fmt.Sprintf("%v seconds since last successful connection, performing update of all spaces.", count))
			RegisterWebHooks()
			GetAllSpaces()
			// since we reset everything, show status space to not become missynced
			// with current channel.
			ChangeSpace("status")
			count = 0
		}

		defer conn.Close()
		conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))

		start := time.Now()
		oneByte := make([]byte, 1)
		_, err = conn.Read(oneByte)
		if err != nil {
			AddStatusText("[red]ERROR READ")
			continue
		}
		duration := time.Since(start)

		var durationAsInt64 = int64(duration) / 1000 / 1000
		lag_color := ""
		switch {
		case durationAsInt64 < 200:
			lag_color = "green"
		case durationAsInt64 < 500:
			lag_color = "yellow"
		case durationAsInt64 < 800:
			lag_color = "orange"
		default:
			lag_color = "red"
		}

		win.StatusLag.SetText(fmt.Sprintf("[navy][[white]Lag: [%s]%v[navy]]", lag_color, duration-(duration%time.Millisecond)))
		win.App.Draw()

		ticker := time.NewTicker(10 * time.Second)
		select {
		case <-ticker.C:
			break
		case <-channels.Quit:
			return
		}

	}
}
