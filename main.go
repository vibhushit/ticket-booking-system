package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Event represents an event with tickets.
type Event struct {
	ID      int
	Name    string
	Tickets []Ticket
}

// Ticket represents a ticket with ID and status.
type Ticket struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// var (
// 	eventsMutex sync.Mutex
// 	events      = make(map[int]*Event)
// )

var eventsMutex sync.Mutex
var events = make(map[int]*Event)

func main() {
	// Simulate initial data

	//key- int = value- address
	events[1] = &Event{
		ID:   1,
		Name: "Event 1",
		Tickets: []Ticket{
			{ID: 1, Status: "available"},
			{ID: 3, Status: "available"},
		},
	}

	events[2] = &Event{
		ID:   2,
		Name: "Event 2",
		Tickets: []Ticket{
			{ID: 4, Status: "available"},
			{ID: 5, Status: "available"},
		},
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/event/", eventHandler)

	//handle via go server
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server listening on port:8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Simplified event list for demonstration
	var eventList []*Event
	for _, e := range events {
		fmt.Println(e)
		eventList = append(eventList, e)
	}

	err = tmpl.Execute(w, eventList)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Path[len("/event/"):] // /event/2 -> 2

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	event, ok := events[eventID]
	if !ok {
		http.Error(w, "Event Not Found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		ticketID, _ := strconv.Atoi(r.FormValue("ticket_id"))
		// Simulate booking ticket by changing its status
		eventsMutex.Lock()
		defer eventsMutex.Unlock()
		for i := range event.Tickets {
			if event.Tickets[i].ID == ticketID && event.Tickets[i].Status == "available" {
				event.Tickets[i].Status = "booked"
				break
			}
		}

		// Update the ticket data in the file system
		err := saveTickets(eventID, event.Tickets)
		if err != nil {
			log.Println("Error saving ticket data:", err)
		}
	}

	tmpl, err := template.ParseFiles("templates/event.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, event)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Functions for managing data on the file system
func loadTickets(eventID int) ([]Ticket, error) {
	fileName := "events/event" + strconv.Itoa(eventID) + "/tickets.json"
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tickets []Ticket
	err = json.NewDecoder(file).Decode(&tickets)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func saveTickets(eventID int, tickets []Ticket) error {
	fileName := "events/event" + strconv.Itoa(eventID) + "/tickets.json"
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(tickets)
	if err != nil {
		return err
	}

	return nil
}
