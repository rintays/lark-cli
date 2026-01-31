package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	docxChatAnnouncementPageSize = 200
	docxChatAnnouncementMaxPages = 20
)

type docxChatAnnouncementError struct {
	msg string
}

func (e *docxChatAnnouncementError) Error() string {
	return fmt.Sprintf("get chat announcement failed: %s", e.msg)
}

func (c *Client) GetChatAnnouncement(ctx context.Context, token string, chatID string, userIDType string) (ChatAnnouncement, error) {
	if !c.available() {
		return ChatAnnouncement{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ChatAnnouncement{}, errors.New("tenant access token is required")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return ChatAnnouncement{}, errors.New("chat id is required")
	}
	userIDType = strings.TrimSpace(userIDType)

	announcement, err := c.getChatAnnouncementIM(ctx, tenantToken, chatID, userIDType)
	if err == nil {
		if announcement.AnnouncementType == "" {
			announcement.AnnouncementType = "doc"
		}
		return announcement, nil
	}
	if !isDocxChatAnnouncementError(err) {
		return ChatAnnouncement{}, err
	}
	return c.getChatAnnouncementDocx(ctx, tenantToken, chatID, userIDType)
}

func (c *Client) getChatAnnouncementIM(ctx context.Context, tenantToken string, chatID string, userIDType string) (ChatAnnouncement, error) {
	builder := im.NewGetChatAnnouncementReqBuilder().ChatId(chatID)
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}
	resp, err := c.sdk.Im.V1.ChatAnnouncement.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ChatAnnouncement{}, err
	}
	if resp == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement failed: empty response")
	}
	if !resp.Success() {
		if isDocxChatAnnouncementMessage(resp.Msg) {
			return ChatAnnouncement{}, &docxChatAnnouncementError{msg: resp.Msg}
		}
		return ChatAnnouncement{}, fmt.Errorf("get chat announcement failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement failed: empty data")
	}
	return mapChatAnnouncement(resp.Data), nil
}

func (c *Client) UpdateChatAnnouncement(ctx context.Context, token string, chatID string, revision string, requests []string) error {
	if !c.available() {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return errors.New("chat id is required")
	}
	if strings.TrimSpace(revision) == "" {
		return errors.New("revision is required")
	}
	if len(requests) == 0 {
		return errors.New("requests are required")
	}

	body := im.NewPatchChatAnnouncementReqBodyBuilder().
		Revision(strings.TrimSpace(revision)).
		Requests(requests).
		Build()
	builder := im.NewPatchChatAnnouncementReqBuilder().
		ChatId(chatID).
		Body(body)

	resp, err := c.sdk.Im.V1.ChatAnnouncement.Patch(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("update chat announcement failed: empty response")
	}
	if !resp.Success() {
		return fmt.Errorf("update chat announcement failed: %s", resp.Msg)
	}
	return nil
}

func mapChatAnnouncement(data *im.GetChatAnnouncementRespData) ChatAnnouncement {
	if data == nil {
		return ChatAnnouncement{}
	}
	out := ChatAnnouncement{AnnouncementType: "doc"}
	if data.Content != nil {
		out.Content = *data.Content
	}
	if data.Revision != nil {
		out.Revision = *data.Revision
	}
	if data.CreateTime != nil {
		out.CreateTime = *data.CreateTime
	}
	if data.UpdateTime != nil {
		out.UpdateTime = *data.UpdateTime
	}
	if data.OwnerIdType != nil {
		out.OwnerIDType = *data.OwnerIdType
	}
	if data.OwnerId != nil {
		out.OwnerID = *data.OwnerId
	}
	if data.ModifierIdType != nil {
		out.ModifierIDType = *data.ModifierIdType
	}
	if data.ModifierId != nil {
		out.ModifierID = *data.ModifierId
	}
	return out
}

func (c *Client) getChatAnnouncementDocx(ctx context.Context, tenantToken string, chatID string, userIDType string) (ChatAnnouncement, error) {
	if !c.docxChatAnnouncementSDKAvailable() {
		return ChatAnnouncement{}, errors.New("docx chat announcement sdk unavailable")
	}
	builder := larkdocx.NewGetChatAnnouncementReqBuilder().ChatId(chatID)
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}
	resp, err := c.sdk.Docx.V1.ChatAnnouncement.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ChatAnnouncement{}, err
	}
	if resp == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement (docx) failed: empty response")
	}
	if !resp.Success() {
		return ChatAnnouncement{}, fmt.Errorf("get chat announcement (docx) failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement (docx) failed: empty data")
	}

	announcement, revisionID := mapDocxChatAnnouncement(resp.Data)
	blocks, err := c.listChatAnnouncementBlocksDocx(ctx, tenantToken, chatID, userIDType, revisionID)
	if err != nil {
		return ChatAnnouncement{}, err
	}
	announcement.Blocks = blocks
	return announcement, nil
}

func (c *Client) listChatAnnouncementBlocksDocx(ctx context.Context, tenantToken string, chatID string, userIDType string, revisionID int) ([]*larkdocx.Block, error) {
	if !c.docxChatAnnouncementSDKAvailable() {
		return nil, errors.New("docx chat announcement sdk unavailable")
	}
	pageToken := ""
	blocks := make([]*larkdocx.Block, 0)
	for page := 0; page < docxChatAnnouncementMaxPages; page++ {
		builder := larkdocx.NewListChatAnnouncementBlockReqBuilder().
			ChatId(chatID).
			PageSize(docxChatAnnouncementPageSize)
		if pageToken != "" {
			builder.PageToken(pageToken)
		}
		if revisionID > 0 {
			builder.RevisionId(revisionID)
		}
		if userIDType != "" {
			builder.UserIdType(userIDType)
		}

		resp, err := c.sdk.Docx.V1.ChatAnnouncementBlock.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, errors.New("list chat announcement blocks failed: empty response")
		}
		if !resp.Success() {
			return nil, fmt.Errorf("list chat announcement blocks failed: %s", resp.Msg)
		}
		if resp.Data == nil {
			return blocks, nil
		}
		blocks = append(blocks, resp.Data.Items...)
		hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
		if !hasMore {
			return blocks, nil
		}
		if resp.Data.PageToken == nil || *resp.Data.PageToken == "" {
			return blocks, nil
		}
		pageToken = *resp.Data.PageToken
	}
	return blocks, errors.New("list chat announcement blocks failed: pagination limit exceeded")
}

func (c *Client) docxChatAnnouncementSDKAvailable() bool {
	return c != nil &&
		c.sdk != nil &&
		c.sdk.Docx != nil &&
		c.sdk.Docx.V1 != nil &&
		c.sdk.Docx.V1.ChatAnnouncement != nil &&
		c.sdk.Docx.V1.ChatAnnouncementBlock != nil
}

func mapDocxChatAnnouncement(data *larkdocx.GetChatAnnouncementRespData) (ChatAnnouncement, int) {
	if data == nil {
		return ChatAnnouncement{}, 0
	}
	out := ChatAnnouncement{}
	revisionID := 0
	if data.RevisionId != nil {
		revisionID = *data.RevisionId
		out.RevisionID = RevisionID(fmt.Sprintf("%d", revisionID))
	}
	if data.AnnouncementType != nil {
		out.AnnouncementType = *data.AnnouncementType
	}
	out.CreateTime = docxTimeString(data.CreateTimeV2, data.CreateTime)
	out.UpdateTime = docxTimeString(data.UpdateTimeV2, data.UpdateTime)
	if data.OwnerIdType != nil {
		out.OwnerIDType = *data.OwnerIdType
	}
	if data.OwnerId != nil {
		out.OwnerID = *data.OwnerId
	}
	if data.ModifierIdType != nil {
		out.ModifierIDType = *data.ModifierIdType
	}
	if data.ModifierId != nil {
		out.ModifierID = *data.ModifierId
	}
	return out, revisionID
}

func docxTimeString(v2 *string, legacy *int64) string {
	if v2 != nil {
		value := strings.TrimSpace(*v2)
		if value != "" {
			return value
		}
	}
	if legacy != nil {
		return fmt.Sprintf("%d", *legacy)
	}
	return ""
}

func isDocxChatAnnouncementMessage(msg string) bool {
	value := strings.ToLower(strings.TrimSpace(msg))
	return strings.Contains(value, "docx") && strings.Contains(value, "announcement")
}

func isDocxChatAnnouncementError(err error) bool {
	var target *docxChatAnnouncementError
	return errors.As(err, &target)
}
