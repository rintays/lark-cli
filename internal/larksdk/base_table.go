package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

func (c *Client) CreateBaseTable(ctx context.Context, token, appToken, tableName, viewName string) (BaseTable, error) {
	if !c.available() {
		return BaseTable{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseTable{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseTable{}, errors.New("app token is required")
	}
	if tableName == "" {
		return BaseTable{}, errors.New("table name is required")
	}

	tableBuilder := larkbitable.NewReqTableBuilder().Name(tableName)
	if viewName != "" {
		tableBuilder.DefaultViewName(viewName)
	}
	body := larkbitable.NewCreateAppTableReqBodyBuilder().Table(tableBuilder.Build()).Build()
	req := larkbitable.NewCreateAppTableReqBuilder().AppToken(appToken).Body(body).Build()
	resp, err := c.sdk.Bitable.V1.AppTable.Create(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseTable{}, err
	}
	if resp == nil {
		return BaseTable{}, errors.New("create base table failed: empty response")
	}
	if !resp.Success() {
		return BaseTable{}, fmt.Errorf("create base table failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return BaseTable{}, nil
	}
	result := BaseTable{}
	if resp.Data.TableId != nil {
		result.TableID = *resp.Data.TableId
	}
	result.Name = tableName
	return result, nil
}

func (c *Client) DeleteBaseTable(ctx context.Context, token, appToken, tableID string) (BaseTableDeleteResult, error) {
	if !c.available() {
		return BaseTableDeleteResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseTableDeleteResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseTableDeleteResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseTableDeleteResult{}, errors.New("table id is required")
	}

	req := larkbitable.NewDeleteAppTableReqBuilder().AppToken(appToken).TableId(tableID).Build()
	resp, err := c.sdk.Bitable.V1.AppTable.Delete(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseTableDeleteResult{}, err
	}
	if resp == nil {
		return BaseTableDeleteResult{}, errors.New("delete base table failed: empty response")
	}
	if !resp.Success() {
		return BaseTableDeleteResult{}, fmt.Errorf("delete base table failed: %s", resp.Msg)
	}
	result := BaseTableDeleteResult{TableID: tableID, Deleted: true}
	return result, nil
}
