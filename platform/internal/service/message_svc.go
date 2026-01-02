package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tgo/captain/platform/internal/model"
)

type MessageService struct {
	httpClient *http.Client
}

func NewMessageService() *MessageService {
	return &MessageService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type IngestRequest struct {
	PlatformAPIKey string                 `json:"platform_api_key"`
	SourceType     string                 `json:"source_type"`
	MessageID      string                 `json:"message_id"`
	FromUser       string                 `json:"from_user"`
	MsgType        string                 `json:"msg_type"`
	Content        string                 `json:"content"`
	RawPayload     map[string]interface{} `json:"raw_payload"`
}

type SendMessageRequest struct {
	PlatformAPIKey string                 `json:"platform_api_key"`
	FromUID        string                 `json:"from_uid"`
	ChannelID      string                 `json:"channel_id"`
	ChannelType    int                    `json:"channel_type"`
	Payload        map[string]interface{} `json:"payload"`
	ClientMsgNo    string                 `json:"client_msg_no"`
}

type SendMessageResult struct {
	OK          bool   `json:"ok"`
	ClientMsgNo string `json:"client_msg_no,omitempty"`
	MessageID   string `json:"message_id,omitempty"`
	Message     string `json:"message,omitempty"`
}

// ProcessInbound handles incoming messages from various platforms
func (s *MessageService) ProcessInbound(ctx context.Context, platform *model.Platform, req interface{}) error {
	// TODO: Implement message normalization and dispatch to TGO API
	// 1. Normalize message format
	// 2. Dispatch to appropriate handler based on platform type
	// 3. Forward to TGO API for processing
	return nil
}

// SendOutbound sends messages to third-party platforms
func (s *MessageService) SendOutbound(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	platformType := string(platform.Type)

	switch platformType {
	case "wecom":
		return s.sendWeComMessage(ctx, platform, req)
	case "feishu":
		return s.sendFeishuMessage(ctx, platform, req)
	case "dingtalk":
		return s.sendDingTalkMessage(ctx, platform, req)
	case "email":
		return s.sendEmailMessage(ctx, platform, req)
	case "custom":
		return s.sendCustomMessage(ctx, platform, req)
	default:
		return nil, fmt.Errorf("unsupported platform type: %s", platformType)
	}
}

func (s *MessageService) sendWeComMessage(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	// Get credentials from platform config
	corpID := platform.Config["corp_id"]
	agentID := platform.Config["agent_id"]
	corpSecret := platform.Config["corp_secret"]

	if corpID == nil || corpSecret == nil {
		return nil, fmt.Errorf("missing WeCom credentials")
	}

	// Get access token
	tokenURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, corpSecret)
	tokenResp, err := s.httpClient.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get WeCom token: %w", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	if tokenResult.ErrCode != 0 {
		return nil, fmt.Errorf("WeCom token error: %s", tokenResult.ErrMsg)
	}

	// Build message payload
	content, _ := req.Payload["content"].(string)
	msgBody := map[string]interface{}{
		"touser":  req.ChannelID,
		"msgtype": "text",
		"agentid": agentID,
		"text": map[string]string{
			"content": content,
		},
	}

	msgData, _ := json.Marshal(msgBody)
	sendURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", tokenResult.AccessToken)
	sendResp, err := s.httpClient.Post(sendURL, "application/json", bytes.NewReader(msgData))
	if err != nil {
		return nil, fmt.Errorf("failed to send WeCom message: %w", err)
	}
	defer sendResp.Body.Close()

	var sendResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MsgID   string `json:"msgid"`
	}
	if err := json.NewDecoder(sendResp.Body).Decode(&sendResult); err != nil {
		return nil, fmt.Errorf("failed to parse send response: %w", err)
	}

	return &SendMessageResult{
		OK:          sendResult.ErrCode == 0,
		ClientMsgNo: req.ClientMsgNo,
		MessageID:   sendResult.MsgID,
		Message:     sendResult.ErrMsg,
	}, nil
}

func (s *MessageService) sendFeishuMessage(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	// Get credentials from platform config
	appID := platform.Config["app_id"]
	appSecret := platform.Config["app_secret"]

	if appID == nil || appSecret == nil {
		return nil, fmt.Errorf("missing Feishu credentials")
	}

	// Get tenant access token
	tokenBody, _ := json.Marshal(map[string]interface{}{
		"app_id":     appID,
		"app_secret": appSecret,
	})
	tokenResp, err := s.httpClient.Post(
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewReader(tokenBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Feishu token: %w", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	if tokenResult.Code != 0 {
		return nil, fmt.Errorf("Feishu token error: %s", tokenResult.Msg)
	}

	// Build message payload
	content, _ := req.Payload["content"].(string)
	msgBody := map[string]interface{}{
		"receive_id": req.ChannelID,
		"msg_type":   "text",
		"content":    fmt.Sprintf(`{"text":"%s"}`, content),
	}

	msgData, _ := json.Marshal(msgBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST",
		"https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id",
		bytes.NewReader(msgData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+tokenResult.TenantAccessToken)

	sendResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send Feishu message: %w", err)
	}
	defer sendResp.Body.Close()

	var sendResult struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(sendResp.Body).Decode(&sendResult); err != nil {
		return nil, fmt.Errorf("failed to parse send response: %w", err)
	}

	return &SendMessageResult{
		OK:          sendResult.Code == 0,
		ClientMsgNo: req.ClientMsgNo,
		MessageID:   sendResult.Data.MessageID,
		Message:     sendResult.Msg,
	}, nil
}

func (s *MessageService) sendDingTalkMessage(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	// Get credentials from platform config
	appKey := platform.Config["app_key"]
	appSecret := platform.Config["app_secret"]
	agentID := platform.Config["agent_id"]

	if appKey == nil || appSecret == nil {
		return nil, fmt.Errorf("missing DingTalk credentials")
	}

	// Get access token
	tokenURL := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", appKey, appSecret)
	tokenResp, err := s.httpClient.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get DingTalk token: %w", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	if tokenResult.ErrCode != 0 {
		return nil, fmt.Errorf("DingTalk token error: %s", tokenResult.ErrMsg)
	}

	// Build message payload
	content, _ := req.Payload["content"].(string)
	msgBody := map[string]interface{}{
		"agent_id":    agentID,
		"userid_list": req.ChannelID,
		"msg": map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		},
	}

	msgData, _ := json.Marshal(msgBody)
	sendURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", tokenResult.AccessToken)
	sendResp, err := s.httpClient.Post(sendURL, "application/json", bytes.NewReader(msgData))
	if err != nil {
		return nil, fmt.Errorf("failed to send DingTalk message: %w", err)
	}
	defer sendResp.Body.Close()

	var sendResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id"`
	}
	if err := json.NewDecoder(sendResp.Body).Decode(&sendResult); err != nil {
		return nil, fmt.Errorf("failed to parse send response: %w", err)
	}

	return &SendMessageResult{
		OK:          sendResult.ErrCode == 0,
		ClientMsgNo: req.ClientMsgNo,
		MessageID:   fmt.Sprintf("%d", sendResult.TaskID),
		Message:     sendResult.ErrMsg,
	}, nil
}

func (s *MessageService) sendEmailMessage(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	// Email sending would typically use SMTP or an email service API
	// For now, return a placeholder indicating this needs SMTP configuration
	smtpHost := platform.Config["smtp_host"]
	if smtpHost == nil {
		return nil, fmt.Errorf("SMTP not configured for this platform")
	}

	// Email implementation would go here using net/smtp or a third-party library
	// This is a placeholder that indicates the feature is available but needs setup
	return &SendMessageResult{
		OK:          false,
		ClientMsgNo: req.ClientMsgNo,
		Message:     "Email sending requires SMTP configuration",
	}, nil
}

func (s *MessageService) sendCustomMessage(ctx context.Context, platform *model.Platform, req *SendMessageRequest) (*SendMessageResult, error) {
	// Custom platform uses a callback URL to send messages
	callbackURL, ok := platform.Config["callback_url"].(string)
	if !ok || callbackURL == "" {
		return nil, fmt.Errorf("callback_url not configured for custom platform")
	}

	// Build request payload
	payload := map[string]interface{}{
		"from_uid":      req.FromUID,
		"channel_id":    req.ChannelID,
		"channel_type":  req.ChannelType,
		"payload":       req.Payload,
		"client_msg_no": req.ClientMsgNo,
	}

	data, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", callbackURL, bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")

	// Add authentication if configured
	if authHeader, ok := platform.Config["auth_header"].(string); ok && authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send to custom platform: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return &SendMessageResult{
			OK:          false,
			ClientMsgNo: req.ClientMsgNo,
			Message:     fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Try to parse response for message_id
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	messageID, _ := result["message_id"].(string)

	return &SendMessageResult{
		OK:          true,
		ClientMsgNo: req.ClientMsgNo,
		MessageID:   messageID,
		Message:     "Message sent successfully",
	}, nil
}
