package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type LeagueData struct {
	Teams      []*Team      `json:"teams"`
	SubLeagues []*SubLeague `json:"subleagues"`
	Divisions  []*Division  `json:"divisions"`
	Leagues    []*League    `json:"leagues"`
}

type Team struct {
	Lineup              []string `json:"lineup"`   // list of player ID strings
	Rotation            []string `json:"rotation"` // list of player ID strings
	Bullpen             []string `json:"bullpen"`  // list of player ID strings
	Bench               []string `json:"bench"`    // list of player ID strings
	SeasonAttributes    []string `json:"seasonAttributes"`
	PermanentAttributes []string `json:"permanentAttributes"`

	ID             string `json:"_id"`
	FullName       string `json:"fullName"`
	Location       string `json:"location"`
	MainColor      string `json:"mainColor"`
	Nickname       string `json:"nickname"`
	SecondaryColor string `json:"secondaryColor"`
	Shorthand      string `json:"shorthand"`
	Emoji          string `json:"emoji"`
	Slogan         string `json:"slogan"`

	ShameRuns      int `json:"shameRuns"`
	TotalShames    int `json:"totalShames"`
	TotalShamings  int `json:"totalShamings"`
	SeasonShames   int `json:"seasonShames"`
	SeasonShamings int `json:"seasonShamings"`
	Championships  int `json:"championships"`
}

type SubLeague struct {
	Divisions []string `json:"divisions"` // list of division ID strings
	ID        string   `json:"_id"`
	Name      string   `json:"name"`
}

type Division struct {
	Teams []string `json:"teams"` // list of team ID strings
	ID    string   `json:"_id"`
	Name  string   `json:"name"`
}

type League struct {
	Subleagues  []string `json:"subleagues"` // list of SubLeague ID strings
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Tiebreakers string   `json:"tiebreakers"`
}

type GameData struct {
	Sim                *Sim                `json:"sim"`
	Season             *Season             `json:"season"`
	Standsings         *Standings          `json:"standings"`
	Schedules          []*Schedule         `json:"schedule"`
	TomorrowsSchedules []*TomorrowSchedule `json:"tomorrowSchedule"`
	PostSeason         *PostSeason         `json:"postseason"`
}

type Sim struct {
	ID              string    `json:"_id"`
	V               int       `json:"__v"`
	Day             int       `json:"day"`
	League          string    `json:"league"`
	NextElectionEnd time.Time `json:"nextElectionEnd"`
	NextPhaseTime   time.Time `json:"nextPhaseTime"`
	NextSeasonStart time.Time `json:"nextSeasonStart"`
	Phase           int       `json:"phase"`
	PlayOffRound    int       `json:"playOffRound"`
	Playoffs        string    `json:"playoffs"`
	Rules           string    `json:"rules"`
	Season          int       `json:"season"`
	SeasonID        string    `json:"seasonId"`
	Terminology     string    `json:"terminology"`
	EraColor        string    `json:"eraColor"`
	EraTitle        string    `json:"eraTitle"`
	OpenedBook      bool      `json:"openedBook"`
	UnlockedPeanuts bool      `json:"unlockedPeanuts"`
	Twgo            string    `json:"twgo"`
	DoTheThing      bool      `json:"doTheThing"`
	LabourOne       int       `json:"labourOne"`
	SubEraColor     string    `json:"subEraColor"`
	SubEraTitle     string    `json:"subEraTitle"`
}

type Season struct {
	ID           string `json:"_id"`
	V            int    `json:"__v"`
	League       string `json:"league"`
	Rules        string `json:"rules"`
	Schedule     string `json:"schedule"`
	SeasonNumber int    `json:"seasonNumber"`
	Standings    string `json:"standings"`
	Stats        string `json:"stats"`
	Terminology  string `json:"terminology"`
}

type Standings struct {
	ID     string         `json:"_id"`
	V      int            `json:"__v"`
	Losses map[string]int `json:"losses"`
	Wins   map[string]int `json:"wins"`
}

type Schedule struct {
	BasesOccupied       []int         `json:"basesOccupied"`
	BaseRunners         []string      `json:"baseRunners"` // list of player ID strings
	Outcomes            []interface{} `json:"outcomes"`    // no idea what this is...
	ID                  string        `json:"_id"`
	Terminology         string        `json:"terminology"`
	LastUpdate          string        `json:"lastUpdate"`
	Rules               string        `json:"rules"`
	Statsheet           string        `json:"statsheet"`
	AwayPitcher         string        `json:"awayPitcher"`
	AwayPitcherName     string        `json:"awayPitcherName"`
	AwayBatter          string        `json:"awayBatter"`
	AwayBatterName      string        `json:"awayBatterName"`
	AwayTeam            string        `json:"awayTeam"`
	AwayTeamName        string        `json:"awayTeamName"`
	AwayTeamNickname    string        `json:"awayTeamNickname"`
	AwayTeamColor       string        `json:"awayTeamColor"`
	AwayTeamEmoji       string        `json:"awayTeamEmoji"`
	AwayOdds            float64       `json:"awayOdds"`
	AwayStrikes         int           `json:"awayStrikes"`
	AwayScore           int           `json:"awayScore"`
	AwayTeamBatterCount int           `json:"awayTeamBatterCount"`
	HomePitcher         string        `json:"homePitcher"`
	HomePitcherName     string        `json:"homePitcherName"`
	HomeBatter          string        `json:"homeBatter"`
	HomeBatterName      string        `json:"homeBatterName"`
	HomeTeam            string        `json:"homeTeam"`
	HomeTeamName        string        `json:"homeTeamName"`
	HomeTeamNickname    string        `json:"homeTeamNickname"`
	HomeTeamColor       string        `json:"homeTeamColor"`
	HomeTeamEmoji       string        `json:"homeTeamEmoji"`
	HomeOdds            float64       `json:"homeOdds"`
	HomeStrikes         int           `json:"homeStrikes"`
	HomeScore           int           `json:"homeScore"`
	HomeTeamBatterCount int           `json:"homeTeamBatterCount"`
	Season              int           `json:"season"`
	IsPostseason        bool          `json:"isPostseason"`
	Day                 int           `json:"day"`
	Phase               int           `json:"phase"`
	GameComplete        bool          `json:"gameComplete"`
	Finalized           bool          `json:"finalized"`
	GameStart           bool          `json:"gameStart"`
	HalfInningOuts      int           `json:"halfInningOuts"`
	HalfInningScore     int           `json:"halfInningScore"`
	Inning              int           `json:"inning"`
	TopOfInning         bool          `json:"topOfInning"`
	AtBatBalls          int           `json:"atBatBalls"`
	AtBatStrikes        int           `json:"atBatStrikes"`
	SeriesIndex         int           `json:"seriesIndex"`
	SeriesLength        int           `json:"seriesLength"`
	Shame               bool          `json:"shame"`
	Weather             int           `json:"weather"`
	BaserunnerCount     int           `json:"baserunnerCount"`
}

type TomorrowSchedule struct {
	BasesOccupied       []int         `json:"basesOccupied"`
	BaseRunners         []string      `json:"baseRunners"` // list of player ID strings
	Outcomes            []interface{} `json:"outcomes"`    // not sure what this contains...
	ID                  string        `json:"_id"`
	Terminology         string        `json:"terminology"`
	LastUpdate          string        `json:"lastUpdate"`
	Rules               string        `json:"rules"`
	Statsheet           string        `json:"statsheet"`
	AwayPitcher         string        `json:"awayPitcher"`
	AwayPitcherName     string        `json:"awayPitcherName"`
	AwayBatter          string        `json:"awayBatter"`
	AwayBatterName      string        `json:"awayBatterName"`
	AwayTeam            string        `json:"awayTeam"`
	AwayTeamName        string        `json:"awayTeamName"`
	AwayTeamNickname    string        `json:"awayTeamNickname"`
	AwayTeamColor       string        `json:"awayTeamColor"`
	AwayTeamEmoji       string        `json:"awayTeamEmoji"`
	AwayOdds            float64       `json:"awayOdds"`
	AwayStrikes         int           `json:"awayStrikes"`
	AwayScore           int           `json:"awayScore"`
	AwayTeamBatterCount int           `json:"awayTeamBatterCount"`
	HomePitcher         string        `json:"homePitcher"`
	HomePitcherName     string        `json:"homePitcherName"`
	HomeBatter          string        `json:"homeBatter"`
	HomeBatterName      string        `json:"homeBatterName"`
	HomeTeam            string        `json:"homeTeam"`
	HomeTeamName        string        `json:"homeTeamName"`
	HomeTeamNickname    string        `json:"homeTeamNickname"`
	HomeTeamColor       string        `json:"homeTeamColor"`
	HomeTeamEmoji       string        `json:"homeTeamEmoji"`
	HomeOdds            float64       `json:"homeOdds"`
	HomeStrikes         int           `json:"homeStrikes"`
	HomeScore           int           `json:"homeScore"`
	HomeTeamBatterCount int           `json:"homeTeamBatterCount"`
	Season              int           `json:"season"`
	IsPostseason        bool          `json:"isPostseason"`
	Day                 int           `json:"day"`
	Phase               int           `json:"phase"`
	GameComplete        bool          `json:"gameComplete"`
	Finalized           bool          `json:"finalized"`
	GameStart           bool          `json:"gameStart"`
	HalfInningOuts      int           `json:"halfInningOuts"`
	HalfInningScore     int           `json:"halfInningScore"`
	Inning              int           `json:"inning"`
	TopOfInning         bool          `json:"topOfInning"`
	AtBatBalls          int           `json:"atBatBalls"`
	AtBatStrikes        int           `json:"atBatStrikes"`
	SeriesIndex         int           `json:"seriesIndex"`
	SeriesLength        int           `json:"seriesLength"`
	Shame               bool          `json:"shame"`
	Weather             int           `json:"weather"`
	BaserunnerCount     int           `json:"baserunnerCount"`
}

type PostSeason struct {
	Playoffs interface{} `json:"playoffs"` // no idea what this is
}

type BlaseballWebSocket struct {
	conn *websocket.Conn
}

func blaseballConnect(ctx context.Context, cookie string) (*BlaseballWebSocket, error) {
	header := http.Header{}
	header.Set("Cookie", cookie)

	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, "wss://blaseball.com/socket.io/?EIO=3&transport=websocket", header)
	if err != nil {
		fmt.Println(resp)
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(body))
		}
		return nil, err
	}

	return &BlaseballWebSocket{conn}, nil
}

func (c *BlaseballWebSocket) Close() error {
	return c.conn.Close()
}

func (c *BlaseballWebSocket) NextUpdate() (interface{}, error) {
	// Get next raw update byte slice
	t, msg, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	} else if t != websocket.TextMessage {
		return nil, errors.New("unsupported websocket message type")
	}

	// Trim initial bytes
	msg = bytes.TrimPrefix(msg, []byte("42["))

	// Get old length before trim
	oldLen := len(msg)

	// Check if we have league data
	if msg = bytes.TrimPrefix(msg, []byte(leagueDataPrefix)); len(msg) == oldLen-len(leagueDataPrefix) {
		// Trim end bytes
		msg = bytes.TrimSuffix(msg, []byte("]"))

		// Try decode data
		leagueData := &LeagueData{}
		err := json.Unmarshal(msg, leagueData)
		if err != nil {
			return nil, err
		}

		return leagueData, err
	}

	// Check if we have game data
	if msg = bytes.TrimPrefix(msg, []byte(gameDataPrefix)); len(msg) == oldLen-len(gameDataPrefix) {
		// Trim end bytes
		msg = bytes.TrimSuffix(msg, []byte("]"))

		// Try decode data
		gameData := &GameData{}
		err := json.Unmarshal(msg, gameData)
		if err != nil {
			return nil, err
		}

		return gameData, nil
	}

	return nil, errors.New("unrecognized websocket data")
}

const (
	leagueDataPrefix = `"leagueDataUpdate",`
	gameDataPrefix   = `"gameDataUpdate",`
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <cookie>\n", os.Args[0])
		os.Exit(1)
	}

	// Create maps
	teamsMap := make(map[string]*Team)
	subLeaguesMap := make(map[string]*SubLeague)
	divisionsMap := make(map[string]*Division)
	leaguesMap := make(map[string]*League)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Connecting to blaseball websocket...")
	conn, err := blaseballConnect(ctx, os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Printf("Connected!\n\n")

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		for {
			fmt.Println("Waiting for update...")
			update, err := conn.NextUpdate()
			if err != nil {
				fmt.Println(err.Error() + "\n")
				if websocket.IsCloseError(err) {
					return
				}
				continue
			}

			fmt.Printf("Update type: ")
			switch update.(type) {
			case *LeagueData:
				fmt.Printf("LeagueData\n")
				data := update.(*LeagueData)
				for _, team := range data.Teams {
					teamsMap[team.ID] = team
				}
				for _, subLeague := range data.SubLeagues {
					subLeaguesMap[subLeague.ID] = subLeague
				}
				for _, division := range data.Divisions {
					divisionsMap[division.ID] = division
				}
				for _, league := range data.Leagues {
					leaguesMap[league.ID] = league
				}

			case *GameData:
				fmt.Printf("GameData\n")

			default:
				fmt.Printf("unknown\n")
			}

			fmt.Println()
		}
	}()

	sig := <-signals
	fmt.Println("Signal received:", sig)
}
