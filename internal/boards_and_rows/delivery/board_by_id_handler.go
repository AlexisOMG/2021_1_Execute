package delivery

import (
	"2021_1_Execute/internal/domain"
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type GetBoardByIDResponce struct {
	Board getBoardByIDResponceContent `json:"board"`
}
type getBoardByIDResponceContent struct {
	ID          int               `json:"id"`
	Access      string            `json:"access"`
	IsStared    bool              `json:"isStared"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Users       boardUsers        `json:"users"`
	Rows        map[int]boardRows `json:"rows"`
}
type boardUser struct {
	ID     int    `json:"id"`
	Avatar string `json:"avatar"`
}
type boardUsers struct {
	Owner   boardUser   `json:"owner, omitempty"`
	Admins  []boardUser `json:"admins, omitempty"`
	Members []boardUser `json:"members, omitempty"`
}
type boardRows struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Position int               `json:"position"`
	Tasks    map[int]boardTask `json:"tasks"`
}
type boardTask struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func BoardToGetResponce(board domain.FullBoardInfo) GetBoardByIDResponce {
	boardUsers := boardUsers{
		Owner:   boardUser{ID: board.Owner.ID, Avatar: board.Owner.Avatar},
		Admins:  []boardUser{},
		Members: []boardUser{},
	}

	rows := make(map[int]boardRows)
	for _, row := range board.Rows {
		tasks := make(map[int]boardTask)
		for _, task := range row.Tasks {
			tasks[task.Position] = boardTask{
				ID:       task.ID,
				Name:     task.Name,
				Position: task.Position,
			}
		}
		rows[row.Position] = boardRows{
			ID:       row.ID,
			Name:     row.Name,
			Position: row.Position,
			Tasks:    tasks,
		}
	}

	content := getBoardByIDResponceContent{
		ID:          board.ID,
		Access:      "",
		IsStared:    false,
		Name:        board.Name,
		Description: board.Description,
		Rows:        rows,
		Users:       boardUsers,
	}
	return GetBoardByIDResponce{Board: content}
}

func (handler *BoardsHandler) GetBoardByID(c echo.Context) error {
	userID, err := handler.sessionHD.IsAuthorized(c)
	if err != nil {
		return err
	}

	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.Wrap(domain.ForbiddenError, "ID should be int")
	}

	ctx := context.Background()
	board, err := handler.boardUC.GetFullBoardInfo(ctx, boardID, userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, BoardToGetResponce(board))

}

type patchBoardByIDRequest struct {
	Access      string     `json:"access,omitempty"`
	IsStared    bool       `json:"isStared,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Users       boardUsers `json:"users,omitempty"`
	Move        rowsMove   `json:"move,omitempty"`
}
type rowsMove struct {
	RowID       int `json:"row_id"`
	NewPosition int `json:"new_position"`
}

func (handler *BoardsHandler) PatchBoardByID(c echo.Context) error {
	userID, err := handler.sessionHD.IsAuthorized(c)
	if err != nil {
		return err
	}

	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.Wrap(domain.ForbiddenError, "ID should be int")
	}
	input := new(patchBoardByIDRequest)
	input.Move.NewPosition = -1
	input.Move.RowID = -1
	if err := c.Bind(input); err != nil {
		return errors.Wrap(domain.BadRequestError, err.Error())
	}

	if input.Name != "" || input.Description != "" {
		err = handler.boardUC.UpdateBoard(context.Background(), domain.Board{ID: boardID, Name: input.Name, Description: input.Description}, userID)
		if err != nil {
			return err
		}
	}
	if input.Move.RowID >= 0 || input.Move.NewPosition >= 0 {
		if !(input.Move.RowID >= 0 && input.Move.NewPosition >= 0) {
			return domain.BadRequestError
		}
		err = handler.boardUC.MoveRow(context.Background(), boardID, input.Move.RowID, input.Move.NewPosition, userID)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (handler *BoardsHandler) DeleteBoardByID(c echo.Context) error {
	userID, err := handler.sessionHD.IsAuthorized(c)
	if err != nil {
		return err
	}

	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.Wrap(domain.ForbiddenError, "ID should be int")
	}
	ctx := context.Background()
	err = handler.boardUC.DeleteBoard(ctx, boardID, userID)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}