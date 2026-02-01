package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	contact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
)

func (c *Client) BatchGetUserIDs(ctx context.Context, token string, req BatchGetUserIDRequest) ([]User, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}

	bodyBuilder := contact.NewBatchGetIdUserReqBodyBuilder()
	if len(req.Emails) > 0 {
		bodyBuilder.Emails(req.Emails)
	}
	if len(req.Mobiles) > 0 {
		bodyBuilder.Mobiles(req.Mobiles)
	}
	builder := contact.NewBatchGetIdUserReqBuilder().Body(bodyBuilder.Build())

	resp, err := c.sdk.Contact.V3.User.BatchGetId(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch get user ids failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("batch get user ids failed: %s", resp.Msg)
	}

	users := []User{}
	if resp.Data != nil && resp.Data.UserList != nil {
		users = make([]User, 0, len(resp.Data.UserList))
		for _, user := range resp.Data.UserList {
			users = append(users, mapContactInfo(user))
		}
	}
	return users, nil
}

func (c *Client) ListUsersByDepartment(ctx context.Context, token string, req ListUsersByDepartmentRequest) (ListUsersByDepartmentResult, error) {
	if !c.available() {
		return ListUsersByDepartmentResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListUsersByDepartmentResult{}, errors.New("tenant access token is required")
	}

	builder := contact.NewFindByDepartmentUserReqBuilder()
	if req.DepartmentID != "" {
		builder.DepartmentId(req.DepartmentID)
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Contact.V3.User.FindByDepartment(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	if resp == nil {
		return ListUsersByDepartmentResult{}, errors.New("list users failed: empty response")
	}
	if !resp.Success() {
		return ListUsersByDepartmentResult{}, fmt.Errorf("list users failed: %s", resp.Msg)
	}

	result := ListUsersByDepartmentResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]User, 0, len(resp.Data.Items))
			for _, user := range resp.Data.Items {
				result.Items = append(result.Items, mapContactUser(user))
			}
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

func (c *Client) GetContactUser(ctx context.Context, token string, req GetContactUserRequest) (User, error) {
	if !c.available() {
		return User{}, ErrUnavailable
	}
	if req.UserID == "" {
		return User{}, fmt.Errorf("user id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return User{}, errors.New("tenant access token is required")
	}

	builder := contact.NewGetUserReqBuilder().UserId(req.UserID)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}
	resp, err := c.sdk.Contact.V3.User.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return User{}, err
	}
	if resp == nil {
		return User{}, errors.New("get contact user failed: empty response")
	}
	if !resp.Success() {
		return User{}, fmt.Errorf("get contact user failed: %s", resp.Msg)
	}

	if resp.Data != nil && resp.Data.User != nil {
		return mapContactUser(resp.Data.User), nil
	}
	return User{}, nil
}

func mapContactInfo(user *contact.UserContactInfo) User {
	if user == nil {
		return User{}
	}
	result := User{}
	if user.UserId != nil {
		result.UserID = *user.UserId
	}
	if user.Email != nil {
		result.Email = *user.Email
	}
	if user.Mobile != nil {
		result.Mobile = *user.Mobile
	}
	return result
}

func mapContactUser(user *contact.User) User {
	if user == nil {
		return User{}
	}
	result := User{}
	if user.UserId != nil {
		result.UserID = *user.UserId
	}
	if user.OpenId != nil {
		result.OpenID = *user.OpenId
	}
	if user.Name != nil {
		result.Name = *user.Name
	}
	if user.Email != nil {
		result.Email = *user.Email
	}
	if user.EnterpriseEmail != nil {
		result.EnterpriseEmail = *user.EnterpriseEmail
	}
	if user.Mobile != nil {
		result.Mobile = *user.Mobile
	}
	if user.DepartmentIds != nil {
		result.DepartmentIDs = append([]string{}, user.DepartmentIds...)
	}
	return result
}
