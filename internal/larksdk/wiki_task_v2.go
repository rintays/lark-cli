package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
)

type WikiTaskMoveResult struct {
	Node      *WikiNode `json:"node,omitempty"`
	Status    int       `json:"status"`
	StatusMsg string    `json:"status_msg,omitempty"`
}

type WikiTaskResult struct {
	TaskID     string               `json:"task_id"`
	MoveResult []WikiTaskMoveResult `json:"move_result,omitempty"`
}

type GetWikiTaskRequest struct {
	TaskID   string
	TaskType string
}

func (c *Client) GetWikiTaskV2(ctx context.Context, token string, req GetWikiTaskRequest) (WikiTaskResult, error) {
	if !c.available() {
		return WikiTaskResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiTaskResult{}, errors.New("tenant access token is required")
	}
	if req.TaskID == "" {
		return WikiTaskResult{}, errors.New("task id is required")
	}

	builder := larkwiki.NewGetTaskReqBuilder().TaskId(req.TaskID)
	if req.TaskType != "" {
		builder = builder.TaskType(req.TaskType)
	}

	resp, err := c.sdk.Wiki.V2.Task.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return WikiTaskResult{}, err
	}
	if resp == nil {
		return WikiTaskResult{}, errors.New("wiki task get failed: empty response")
	}
	if !resp.Success() {
		return WikiTaskResult{}, fmt.Errorf("wiki task get failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Task == nil {
		return WikiTaskResult{}, errors.New("wiki task get failed: missing task")
	}

	out := WikiTaskResult{}
	if resp.Data.Task.TaskId != nil {
		out.TaskID = *resp.Data.Task.TaskId
	}
	if resp.Data.Task.MoveResult == nil {
		return out, nil
	}
	out.MoveResult = make([]WikiTaskMoveResult, 0, len(resp.Data.Task.MoveResult))
	for _, mr := range resp.Data.Task.MoveResult {
		if mr == nil {
			continue
		}
		mapped := WikiTaskMoveResult{}
		if mr.Status != nil {
			mapped.Status = *mr.Status
		}
		if mr.StatusMsg != nil {
			mapped.StatusMsg = *mr.StatusMsg
		}
		if mr.Node != nil {
			n := convertWikiNode(mr.Node)
			mapped.Node = &n
		}
		out.MoveResult = append(out.MoveResult, mapped)
	}
	return out, nil
}

func (c *Client) GetWikiTaskV2WithUserToken(ctx context.Context, userAccessToken string, req GetWikiTaskRequest) (WikiTaskResult, error) {
	if !c.available() {
		return WikiTaskResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return WikiTaskResult{}, errors.New("user access token is required")
	}
	if req.TaskID == "" {
		return WikiTaskResult{}, errors.New("task id is required")
	}

	builder := larkwiki.NewGetTaskReqBuilder().TaskId(req.TaskID)
	if req.TaskType != "" {
		builder = builder.TaskType(req.TaskType)
	}

	resp, err := c.sdk.Wiki.V2.Task.Get(ctx, builder.Build(), larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return WikiTaskResult{}, err
	}
	if resp == nil {
		return WikiTaskResult{}, errors.New("wiki task get failed: empty response")
	}
	if !resp.Success() {
		return WikiTaskResult{}, fmt.Errorf("wiki task get failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Task == nil {
		return WikiTaskResult{}, errors.New("wiki task get failed: missing task")
	}

	out := WikiTaskResult{}
	if resp.Data.Task.TaskId != nil {
		out.TaskID = *resp.Data.Task.TaskId
	}
	if resp.Data.Task.MoveResult == nil {
		return out, nil
	}
	out.MoveResult = make([]WikiTaskMoveResult, 0, len(resp.Data.Task.MoveResult))
	for _, mr := range resp.Data.Task.MoveResult {
		if mr == nil {
			continue
		}
		mapped := WikiTaskMoveResult{}
		if mr.Status != nil {
			mapped.Status = *mr.Status
		}
		if mr.StatusMsg != nil {
			mapped.StatusMsg = *mr.StatusMsg
		}
		if mr.Node != nil {
			n := convertWikiNode(mr.Node)
			mapped.Node = &n
		}
		out.MoveResult = append(out.MoveResult, mapped)
	}
	return out, nil
}
