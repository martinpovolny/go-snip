package client

import (
	"encoding/json"
	"fmt"
	"log"
	"martinp/piskvorky/pisk"
	"os"
	"strconv"

	"github.com/apibillme/restly"
)

type APIClient struct {
	urlBase string
}

type User struct {
	UserId    string `json:"userId"`
	UserToken string `json:"userToken"`
}

type Game struct {
	GameId    string `json:"gameId"`
	GameToken string `json:"gameToken"`
}

func (c APIClient) RegisterPlayer(nickname, email string) (*User, error) {
	jsonBody, _ := json.Marshal(map[string]string{
		"nickname": nickname,
		"email":    email,
	})

	res, statusCode, err := restly.PostJSON(restly.New(), c.urlBase+"/api/v1/user", string(jsonBody), "")
	if err != nil {
		return nil, err
	}
	if statusCode != 201 {
		return nil, fmt.Errorf("error registering player: %d", statusCode)
	}

	var user User
	log.Printf("%s", res.Raw)
	json.Unmarshal([]byte(res.Raw), &user)
	return &user, nil
}

func (c APIClient) StartGame(userToken string) (*Game, error) {
	jsonBody, _ := json.Marshal(map[string]string{
		"userToken": userToken,
	})
	log.Println("start game request", string(jsonBody))
	res, statusCode, err := restly.PostJSON(restly.New(), c.urlBase+"/api/v1/connect", string(jsonBody), "")
	if err != nil {
		return nil, err
	}
	if statusCode != 201 {
		log.Printf("%s", res.Raw)
		return nil, fmt.Errorf("error starting a game: %d", statusCode)
	}

	var game Game
	log.Printf("%s", res.Raw)
	json.Unmarshal([]byte(res.Raw), &game)
	log.Println("Game: ", game.GameId, game.GameToken)
	return &game, nil
}

func (c APIClient) Play(userToken, gameToken string, x, y uint) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(map[string]string{
		"userToken": userToken,
		"gameToken": gameToken,
		"positionX": strconv.FormatInt(int64(x), 10),
		"positionY": strconv.FormatInt(int64(y), 10),
	})
	log.Println("play request", string(jsonBody))
	res, statusCode, err := restly.PostJSON(restly.New(), c.urlBase+"/api/v1/play", string(jsonBody), "")
	if err != nil {
		return nil, err
	}
	if statusCode != 201 {
		log.Printf("%s", res.Raw)
		return nil, fmt.Errorf("error playing: %d", statusCode)
	}

	var result map[string]interface{}
	log.Printf("%s", res.Raw)
	json.Unmarshal([]byte(res.Raw), &result)
	return result, nil
}

func (c APIClient) CheckGameStatus(userToken, gameToken string) (int, map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(map[string]string{
		"userToken": userToken,
		"gameToken": gameToken,
	})
	log.Println("game status request", string(jsonBody))
	res, statusCode, err := restly.PostJSON(restly.New(), c.urlBase+"/api/v1/checkStatus", string(jsonBody), "")
	if err != nil {
		return 0, nil, err
	}
	if statusCode != 200 && statusCode != 226 {
		log.Printf("%s", res.Raw)
		return statusCode, nil, fmt.Errorf("error checking status: %d", statusCode)
	}

	var result map[string]interface{}
	log.Printf("%s", res.Raw)
	json.Unmarshal([]byte(res.Raw), &result)
	return statusCode, result, nil
}

type RemotePlay struct {
	apiClient APIClient
	user      User
	game      Game
}

func NewRemotePlay() RemotePlay {
	return RemotePlay{
		APIClient{"https://piskvorky.jobs.cz"},
		User{},
		Game{},
	}
}

func (r *RemotePlay) LoadCredentialsFromConfigFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	configuration := map[string]string{}
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&configuration)
	if err != nil {
		return err
	}

	r.user.UserId = configuration["userId"]
	r.user.UserToken = configuration["userToken"]
	r.game.GameId = configuration["gameId"]
	r.game.GameToken = configuration["gameToken"]

	log.Println("Loaded credentials from config file", r.user.UserId, r.user.UserToken, r.game.GameId, r.game.GameToken)
	return nil
}

func (r *RemotePlay) SetGame(gameId, gameToken string) *RemotePlay {
	r.game.GameId = gameId
	r.game.GameToken = gameToken
	return r
}

/* LoadGame returns true in the first return value if the game is finished.
It returns the Game in the second argument and then an error in the third. */
func (r *RemotePlay) LoadGame() (bool, *pisk.Game, error) {
	code, gameData, err := r.apiClient.CheckGameStatus(r.user.UserToken, r.game.GameToken)
	if err != nil {
		log.Fatal(err)
	}

	coordinates, ok := gameData["coordinates"].([]interface{})
	if !ok {
		return false, nil, fmt.Errorf("coordinates not found in game data")
	}
	return code == 226, newGameFromCoordinates(coordinates), nil
}

func newGameFromCoordinates(cordinates []interface{}) *pisk.Game {
	game := pisk.NewGame(32, true) // asumption: the game is always 32x32 and X always starts
	var player uint8 = 0
	for _, move := range cordinates {
		moveData, ok := move.(map[string]interface{})
		if !ok {
			log.Fatal("Error parsing coordinates: move is not a map", move)
		}
		log.Println(moveData)

		var xf, yf float64
		var okx, oky bool

		xf, okx = moveData["x"].(float64)
		yf, oky = moveData["y"].(float64)

		if !okx || !oky {
			log.Println("Error parsing coordinates", moveData)
		}

		// assumption: the game is always 32x32, the coordinates are -15..16
		game.Play(pisk.Move{X: uint8(xf + 16), Y: uint8(yf + 16)}, player)
		player = 1 - player
	}
	return game
}

func (r *RemotePlay) StartGame() (*Game, error) {
	game, error := r.apiClient.StartGame(r.user.UserToken)
	if error != nil {
		return nil, error
	}
	r.game.GameId = game.GameId
	r.game.GameToken = game.GameToken
	return game, nil
}
