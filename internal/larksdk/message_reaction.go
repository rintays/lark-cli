package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) CreateMessageReaction(ctx context.Context, token string, messageID string, emojiType string) (MessageReaction, error) {
	if !c.available() {
		return MessageReaction{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return MessageReaction{}, errors.New("tenant access token is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return MessageReaction{}, errors.New("message id is required")
	}
	emojiType = strings.TrimSpace(emojiType)
	if emojiType == "" {
		return MessageReaction{}, errors.New("emoji type is required")
	}

	body := im.NewCreateMessageReactionReqBodyBuilder().
		ReactionType(im.NewEmojiBuilder().EmojiType(emojiType).Build()).
		Build()
	builder := im.NewCreateMessageReactionReqBuilder().
		MessageId(messageID).
		Body(body)
	resp, err := c.sdk.Im.V1.MessageReaction.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return MessageReaction{}, err
	}
	if resp == nil {
		return MessageReaction{}, errors.New("create message reaction failed: empty response")
	}
	if !resp.Success() {
		return MessageReaction{}, fmt.Errorf("create message reaction failed: %s", resp.Msg)
	}
	if resp.Data != nil {
		return mapMessageReactionCreate(resp.Data), nil
	}
	return MessageReaction{}, nil
}

func (c *Client) DeleteMessageReaction(ctx context.Context, token string, messageID string, reactionID string) (MessageReaction, error) {
	if !c.available() {
		return MessageReaction{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return MessageReaction{}, errors.New("tenant access token is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return MessageReaction{}, errors.New("message id is required")
	}
	reactionID = strings.TrimSpace(reactionID)
	if reactionID == "" {
		return MessageReaction{}, errors.New("reaction id is required")
	}

	builder := im.NewDeleteMessageReactionReqBuilder().
		MessageId(messageID).
		ReactionId(reactionID)
	resp, err := c.sdk.Im.V1.MessageReaction.Delete(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return MessageReaction{}, err
	}
	if resp == nil {
		return MessageReaction{}, errors.New("delete message reaction failed: empty response")
	}
	if !resp.Success() {
		return MessageReaction{}, fmt.Errorf("delete message reaction failed: %s", resp.Msg)
	}
	if resp.Data != nil {
		return mapMessageReactionDelete(resp.Data), nil
	}
	return MessageReaction{}, nil
}

func mapMessageReactionCreate(reaction *im.CreateMessageReactionRespData) MessageReaction {
	if reaction == nil {
		return MessageReaction{}
	}
	return mapMessageReactionFields(reaction.ReactionId, reaction.Operator, reaction.ActionTime, reaction.ReactionType)
}

func mapMessageReactionDelete(reaction *im.DeleteMessageReactionRespData) MessageReaction {
	if reaction == nil {
		return MessageReaction{}
	}
	return mapMessageReactionFields(reaction.ReactionId, reaction.Operator, reaction.ActionTime, reaction.ReactionType)
}

func mapMessageReactionFields(reactionID *string, operator *im.Operator, actionTime *string, reactionType *im.Emoji) MessageReaction {
	out := MessageReaction{}
	if reactionID != nil {
		out.ReactionID = *reactionID
	}
	if operator != nil {
		out.Operator = mapReactionOperator(operator)
	}
	if actionTime != nil {
		out.ActionTime = *actionTime
	}
	if reactionType != nil {
		out.Reaction = mapEmoji(reactionType)
	}
	return out
}

func mapReactionOperator(operator *im.Operator) ReactionOperator {
	if operator == nil {
		return ReactionOperator{}
	}
	out := ReactionOperator{}
	if operator.OperatorId != nil {
		out.OperatorID = *operator.OperatorId
	}
	if operator.OperatorType != nil {
		out.OperatorType = *operator.OperatorType
	}
	return out
}

func mapEmoji(emoji *im.Emoji) Emoji {
	if emoji == nil {
		return Emoji{}
	}
	out := Emoji{}
	if emoji.EmojiType != nil {
		out.EmojiType = *emoji.EmojiType
	}
	return out
}
