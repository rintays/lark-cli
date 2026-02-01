package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	contact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
)

type SearchUsersRequest struct {
	Query     string
	PageSize  int
	PageToken string
}

type SearchUsersResult struct {
	Users     []User
	PageToken string
	HasMore   bool
}

type userSearchResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *userSearchResponseData `json:"data"`
}

type userSearchResponseData struct {
	Users     []User  `json:"users"`
	PageToken *string `json:"page_token"`
	HasMore   *bool   `json:"has_more"`
}

func (r *userSearchResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchUsers(ctx context.Context, userAccessToken string, req SearchUsersRequest) (SearchUsersResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SearchUsersResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return SearchUsersResult{}, errors.New("user access token is required")
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return SearchUsersResult{}, errors.New("query is required")
	}
	if req.PageSize < 0 {
		return SearchUsersResult{}, errors.New("page_size must be greater than or equal to 0")
	}
	if req.PageSize > 200 {
		return SearchUsersResult{}, errors.New("page_size must be less than or equal to 200")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/search/v1/user",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
	}
	apiReq.QueryParams.Set("query", query)
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprint(req.PageSize))
	}
	if strings.TrimSpace(req.PageToken) != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return SearchUsersResult{}, err
	}
	if apiResp == nil {
		return SearchUsersResult{}, errors.New("search users failed: empty response")
	}
	resp := &userSearchResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SearchUsersResult{}, err
	}
	if !resp.Success() {
		return SearchUsersResult{}, fmt.Errorf("search users failed (code=%d): %s", resp.Code, resp.Msg)
	}

	result := SearchUsersResult{}
	if resp.Data != nil {
		if resp.Data.Users != nil {
			result.Users = resp.Data.Users
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}

	if len(result.Users) == 0 {
		return result, nil
	}
	if err := c.enrichSearchUsers(ctx, userAccessToken, result.Users); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) enrichSearchUsers(ctx context.Context, userAccessToken string, users []User) error {
	if !c.available() {
		return ErrUnavailable
	}
	if len(users) == 0 {
		return nil
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return errors.New("user access token is required")
	}

	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		if user.UserID != "" {
			userIDs = append(userIDs, user.UserID)
		}
	}
	if len(userIDs) == 0 {
		return nil
	}

	userReq := contact.NewBatchUserReqBuilder().
		UserIds(userIDs).
		UserIdType("user_id").
		DepartmentIdType("department_id").
		Build()
	userResp, err := c.sdk.Contact.V3.User.Batch(ctx, userReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return err
	}
	if userResp == nil {
		return errors.New("batch users failed: empty response")
	}
	if !userResp.Success() {
		return fmt.Errorf("batch users failed: %s", userResp.Msg)
	}

	lookup := map[string]*contact.User{}
	if userResp.Data != nil && userResp.Data.Items != nil {
		for _, item := range userResp.Data.Items {
			if item == nil || item.UserId == nil || *item.UserId == "" {
				continue
			}
			lookup[*item.UserId] = item
		}
	}

	departmentIDs := map[string]struct{}{}
	for _, item := range lookup {
		if item.DepartmentIds == nil {
			continue
		}
		for _, deptID := range item.DepartmentIds {
			if deptID == "" {
				continue
			}
			departmentIDs[deptID] = struct{}{}
		}
	}

	departmentNames := map[string]string{}
	if len(departmentIDs) > 0 {
		deptIDs := make([]string, 0, len(departmentIDs))
		for deptID := range departmentIDs {
			deptIDs = append(deptIDs, deptID)
		}
		deptReq := contact.NewBatchDepartmentReqBuilder().
			DepartmentIds(deptIDs).
			DepartmentIdType("department_id").
			Build()
		deptResp, err := c.sdk.Contact.V3.Department.Batch(ctx, deptReq, larkcore.WithUserAccessToken(userAccessToken))
		if err != nil {
			return err
		}
		if deptResp == nil {
			return errors.New("batch departments failed: empty response")
		}
		if !deptResp.Success() {
			return fmt.Errorf("batch departments failed: %s", deptResp.Msg)
		}
		if deptResp.Data != nil && deptResp.Data.Items != nil {
			for _, dept := range deptResp.Data.Items {
				if dept == nil || dept.DepartmentId == nil || *dept.DepartmentId == "" {
					continue
				}
				name := ""
				if dept.Name != nil {
					name = *dept.Name
				}
				if name == "" && dept.I18nName != nil {
					if dept.I18nName.ZhCn != nil && *dept.I18nName.ZhCn != "" {
						name = *dept.I18nName.ZhCn
					} else if dept.I18nName.EnUs != nil && *dept.I18nName.EnUs != "" {
						name = *dept.I18nName.EnUs
					} else if dept.I18nName.JaJp != nil && *dept.I18nName.JaJp != "" {
						name = *dept.I18nName.JaJp
					}
				}
				if name == "" {
					continue
				}
				departmentNames[*dept.DepartmentId] = name
			}
		}
	}

	for i := range users {
		user := &users[i]
		detail, ok := lookup[user.UserID]
		if !ok {
			continue
		}
		if detail.Email != nil && *detail.Email != "" {
			user.Email = *detail.Email
		}
		if detail.EnterpriseEmail != nil && *detail.EnterpriseEmail != "" {
			user.EnterpriseEmail = *detail.EnterpriseEmail
		}
		if detail.DepartmentIds != nil {
			user.DepartmentIDs = append([]string{}, detail.DepartmentIds...)
			user.Departments = make([]DepartmentInfo, 0, len(detail.DepartmentIds))
			for _, deptID := range detail.DepartmentIds {
				if deptID == "" {
					continue
				}
				user.Departments = append(user.Departments, DepartmentInfo{
					ID:   deptID,
					Name: departmentNames[deptID],
				})
			}
		}
	}
	return nil
}
