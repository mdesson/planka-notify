package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type ProjectsResp struct {
	Included struct {
		Boards []struct {
			Id                         string     `json:"id"`
			CreatedAt                  time.Time  `json:"createdAt"`
			UpdatedAt                  *time.Time `json:"updatedAt"`
			Position                   int        `json:"position"`
			Name                       string     `json:"name"`
			DefaultView                string     `json:"defaultView"`
			DefaultCardType            string     `json:"defaultCardType"`
			LimitCardTypesToDefaultOne bool       `json:"limitCardTypesToDefaultOne"`
			AlwaysDisplayCardCreator   bool       `json:"alwaysDisplayCardCreator"`
			DisplayCardAges            bool       `json:"displayCardAges"`
			ExpandTaskListsByDefault   bool       `json:"expandTaskListsByDefault"`
			ProjectId                  string     `json:"projectId"`
		} `json:"boards"`
	} `json:"included"`
}

type Card struct {
	Id                string      `json:"id"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
	Type              string      `json:"type"`
	Position          int         `json:"position"`
	Name              string      `json:"name"`
	Description       *string     `json:"description"`
	DueDate           *time.Time  `json:"dueDate"`
	IsDueCompleted    *bool       `json:"isDueCompleted"`
	Stopwatch         interface{} `json:"stopwatch"`
	CommentsTotal     int         `json:"commentsTotal"`
	IsClosed          bool        `json:"isClosed"`
	ListChangedAt     time.Time   `json:"listChangedAt"`
	BoardId           string      `json:"boardId"`
	ListId            string      `json:"listId"`
	CreatorUserId     string      `json:"creatorUserId"`
	PrevListId        interface{} `json:"prevListId"`
	CoverAttachmentId interface{} `json:"coverAttachmentId"`
	IsSubscribed      bool        `json:"isSubscribed"`
}

type BoardsResp struct {
	Item struct {
		Id                         string    `json:"id"`
		CreatedAt                  time.Time `json:"createdAt"`
		UpdatedAt                  time.Time `json:"updatedAt"`
		Position                   int       `json:"position"`
		Name                       string    `json:"name"`
		DefaultView                string    `json:"defaultView"`
		DefaultCardType            string    `json:"defaultCardType"`
		LimitCardTypesToDefaultOne bool      `json:"limitCardTypesToDefaultOne"`
		AlwaysDisplayCardCreator   bool      `json:"alwaysDisplayCardCreator"`
		DisplayCardAges            bool      `json:"displayCardAges"`
		ExpandTaskListsByDefault   bool      `json:"expandTaskListsByDefault"`
		ProjectId                  string    `json:"projectId"`
		IsSubscribed               bool      `json:"isSubscribed"`
	} `json:"item"`
	Included struct {
		Labels []struct {
			Id        string     `json:"id"`
			CreatedAt time.Time  `json:"createdAt"`
			UpdatedAt *time.Time `json:"updatedAt"`
			Position  int        `json:"position"`
			Name      string     `json:"name"`
			Color     string     `json:"color"`
			BoardId   string     `json:"boardId"`
		} `json:"labels"`
		Lists []struct {
			Id        string      `json:"id"`
			CreatedAt time.Time   `json:"createdAt"`
			UpdatedAt *time.Time  `json:"updatedAt"`
			Type      string      `json:"type"`
			Position  *int        `json:"position"`
			Name      *string     `json:"name"`
			Color     interface{} `json:"color"`
			BoardId   string      `json:"boardId"`
		} `json:"lists"`
		Cards      []Card `json:"cards"`
		CardLabels []struct {
			Id        string      `json:"id"`
			CreatedAt time.Time   `json:"createdAt"`
			UpdatedAt interface{} `json:"updatedAt"`
			CardId    string      `json:"cardId"`
			LabelId   string      `json:"labelId"`
		} `json:"cardLabels"`
		TaskLists []struct {
			Id                 string      `json:"id"`
			CreatedAt          time.Time   `json:"createdAt"`
			UpdatedAt          interface{} `json:"updatedAt"`
			Position           int         `json:"position"`
			Name               string      `json:"name"`
			ShowOnFrontOfCard  bool        `json:"showOnFrontOfCard"`
			HideCompletedTasks bool        `json:"hideCompletedTasks"`
			CardId             string      `json:"cardId"`
		} `json:"taskLists"`
		Tasks []struct {
			Id             string      `json:"id"`
			CreatedAt      time.Time   `json:"createdAt"`
			UpdatedAt      time.Time   `json:"updatedAt"`
			Position       int         `json:"position"`
			Name           string      `json:"name"`
			IsCompleted    bool        `json:"isCompleted"`
			TaskListId     string      `json:"taskListId"`
			LinkedCardId   interface{} `json:"linkedCardId"`
			AssigneeUserId interface{} `json:"assigneeUserId"`
		} `json:"tasks"`
	} `json:"included"`
}

func main() {
	// get projects
	projectsBytes, err := PlankaGet("/projects")
	if err != nil {
		panic(err)
	}

	var projects ProjectsResp
	err = json.Unmarshal(projectsBytes, &projects)
	if err != nil {
		panic(err)
	}

	// get all cards we're interested in
	cards := make([]Card, 0)
	for _, board := range projects.Included.Boards {
		boardDetails, err := PlankaGet("/boards/" + board.Id)
		if err != nil {
			panic(err)
		}
		var boards BoardsResp
		if err := json.Unmarshal(boardDetails, &boards); err != nil {
			panic(err)
		}

		for _, card := range boards.Included.Cards {
			// get cards that are incomplete and due today
			if !card.IsClosed && card.DueDate != nil && sameDay(*card.DueDate, time.Now()) {
				cards = append(cards, card)
			}
		}
	}

	// send email
	for _, card := range cards {
		if err := sendEmail(card); err != nil {
			panic(err)
		}
	}
}

func PlankaGet(path string) ([]byte, error) {
	baseURL := os.Getenv("BASE_URL") + "/api"
	plankaToken := os.Getenv("PLANKA_TOKEN")

	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+plankaToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error %s on planka get: %s", resp.Status, string(body))
	}

	return body, nil
}

func sameDay(t1, t2 time.Time) bool {
	loc, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		panic(err)
	}

	y1, m1, d1 := t1.In(loc).Date()
	y2, m2, d2 := t2.In(loc).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func sendEmail(c Card) error {
	from := os.Getenv("GMAIL_FROM_ADDRESS")
	to := os.Getenv("GMAIL_TO_ADDRESS")
	password := os.Getenv("GMAIL_APP_PASSWORD")
	subject := c.Name
	body := fmt.Sprintf("%s/cards/%s", os.Getenv("BASE_URL"), c.Id)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	return smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
}
