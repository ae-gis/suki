/*  raw_test.go
*
* @Author:             Nanang Suryadi
* @Date:               November 24, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 24/11/19 22:09
 */

package sqlx

import (
        "testing"
        "time"

        "gitlab.com/suryakencana007/suki"

        "github.com/stretchr/testify/assert"
)

func TestRawQuery_Select2Join(t *testing.T) {
        member := &Member{
                ID:        34,
                Name:      "Utomo Abdai",
        }
        user := &User{
                ID:        12,
                Name:      "Budi Utomo",
        }
        game := &Game{
                ID:          507,
                Code:        suki.UUID(),
                Title:       "DOTA2",
                Description: String("Froze"),
        }
        game.Enabled = true
        rawSelect := Build()
        query, args := rawSelect.Select(game).Join(user, "id", "user_id").Join(member, "id", "member_id").ToSQL()
        suki.Info(query)
        assert.Contains(t, query, "SELECT refGame.enabled, refGame.game_code, refGame.game_description, refGame.game_id, refGame.game_title FROM ref_game refGame INNER JOIN ref_user refUser ON refUser.id = refGame.user_id")
        assert.Equal(t, len(args), 5)
}

func TestRawQuery_SelectJoin(t *testing.T) {
        user := &User{
                ID:        12,
                Name:      "Budi Utomo",
        }
        game := &Game{
                ID:          507,
                Code:        suki.UUID(),
                Title:       "DOTA2",
                Description: String("Froze"),
        }
        game.Enabled = true
        rawSelect := Build()
        query, args := rawSelect.Select(game).Join(user, "id", "user_id").Where("name = ?", "budi").ToSQL()
        suki.Info(query)
        assert.Contains(t, query, "SELECT refGame.enabled, refGame.game_code, refGame.game_description, refGame.game_id, refGame.game_title FROM ref_game refGame INNER JOIN ref_user refUser ON refUser.id = refGame.user_id  WHERE name = $6")
        assert.Equal(t, len(args), 6)
}

func TestRawQuery_Select(t *testing.T) {
        game := &Game{
                ID:          507,
                Code:        suki.UUID(),
                Title:       "DOTA2",
                Description: String("Froze"),
        }
        game.Enabled = true
        rawSelect := Build()
        query, args := rawSelect.Select(game).ToSQL()
        assert.Contains(t, query, "SELECT refGame.enabled, refGame.game_code, refGame.game_description, refGame.game_id, refGame.game_title FROM ref_game refGame")
        assert.Equal(t, len(args), 5)
}

func TestRawQuery_UpdatesFewField(t *testing.T) {
        game := &Game{
                ID:          507,
                Code:        suki.UUID(),
                Title:       "DOTA2",
                Description: String("Froze"),
        }
        game.Enabled = true
        rawUpdates := Build()
        query, args := rawUpdates.Updates(game).Where("game_code = ? AND game_description > ?", game.Code, 23).ToSQL()
        t.Log(query)
        assert.Contains(t, query, "UPDATE ref_game SET enabled = $2, game_code = $3, game_description = $4, game_id = $5, game_title = $6, write_date = $1")
        assert.Equal(t, len(args), 8)
}

func TestRawQuery_Updates(t *testing.T) {
        game := newGame()
        rawUpdates := Build()
        query, args := rawUpdates.Updates(game).Where("game_code = ? AND game_description > ?", game.Code, 23).ToSQL()
        assert.Contains(t, query, "UPDATE ref_game SET enabled = $2, game_code = $3, game_description = $4, game_id = $5, game_title = $6, rate = $7, release = $8, write_date = $1")
        assert.Equal(t, len(args), 10)
}

func TestRawQuery_Insert(t *testing.T) {
        game := newGame()
        raw := Build().SetTag("sql")
        query, args := raw.Insert(game).Where("game_code = ? AND game_description > ?", game.Code, 23).ToSQL()
        assert.Contains(t, query, "INSERT INTO ref_game (enabled, game_code, game_description, game_id, game_title, rate, release, create_date, write_date)")
        assert.Equal(t, len(args), 11)
}

func TestRawQuery_Inserts(t *testing.T) {
        games := make([]*Game, 0)
        game := newGame()
        games = append(games, game)
        games = append(games, game)
        games = append(games, game)
        raw := Build().SetTag("sql")
        query, args := raw.Inserts(games).Where("game_code = ? AND game_description > ?", game.Code, 23).ToSQL()
        assert.Contains(t, query, "INSERT INTO ref_game (enabled, game_code, game_description, game_id, game_title, rate, release, create_date, write_date)")
        assert.Equal(t, len(args), 29)
}

func newGame() *Game {
        return &Game{
                ID:          507,
                Code:        suki.UUID(),
                Title:       "DOTA2",
                Description: String("Froze"),
                Enabled:     true,
                Rate:        Int64(75),
                Release:     Time(time.Now().UTC()),
        }
}

type Game struct {
        CreatedAt   time.Time  `json:"create_date,omitempty"`
        CreatedBy   string     `json:"created_by,omitempty"`
        UpdatedAt   time.Time  `json:"write_date,omitempty"`
        UpdatedBy   string     `json:"updated_by,omitempty"`
        DeletedAt   time.Time  `json:"deleted_at,omitempty"`
        ID          int        `json:"game_id" sql:"game_id"`
        Code        string     `json:"game_code" sql:"game_code"`
        Title       string     `json:"game_title" sql:"game_title"`
        Description NullString `json:"game_description" sql:"game_description"`
        Enabled     bool       `json:"enabled" sql:"enabled"`
        Rate        NullInt64  `json:"rate" sql:"rate"`
        Release     NullTime   `json:"release" sql:"release"`
}

func (Game) TableName() string {
        return "ref_game"
}


type User struct {
        CreatedAt   time.Time  `json:"create_date,omitempty"`
        CreatedBy   string     `json:"created_by,omitempty"`
        UpdatedAt   time.Time  `json:"write_date,omitempty"`
        UpdatedBy   string     `json:"updated_by,omitempty"`
        DeletedAt   time.Time  `json:"deleted_at,omitempty"`
        ID          int        `json:"user_id" sql:"id"`
        Name        string     `json:"user_name" sql:"name"`
}

func (User) TableName() string {
        return "ref_user"
}

type Member struct {
        CreatedAt   time.Time  `json:"create_date,omitempty"`
        CreatedBy   string     `json:"created_by,omitempty"`
        UpdatedAt   time.Time  `json:"write_date,omitempty"`
        UpdatedBy   string     `json:"updated_by,omitempty"`
        DeletedAt   time.Time  `json:"deleted_at,omitempty"`
        ID          int        `json:"user_id" sql:"id"`
        Name        string     `json:"user_name" sql:"name"`
}

func (Member) TableName() string {
        return "ref_member"
}
