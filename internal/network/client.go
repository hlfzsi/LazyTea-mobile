package network
import (
	"fmt"
	"net/url"
	"sync"
	"time"
	"lazytea-mobile/internal/utils"
	"github.com/gorilla/websocket"
)
type ConnectionCallback func(connected bool)
type MessageCallback func(header MessageHeader, payload interface{})
type RequestCallback struct {
	Success func(payload interface{})
	Error   func(error)
}
type Client struct {
	conn       *websocket.Conn
	connected  bool
	mutex      sync.RWMutex
	writeMutex sync.Mutex  
	logger     *utils.Logger
	host       string
	port       int
	token      string
	connectionCallbacks []ConnectionCallback
	messageCallbacks    map[string][]MessageCallback
	pendingRequests map[string]*RequestCallback
	requestMutex    sync.RWMutex
	stopCh chan struct{}
	doneCh chan struct{}
	lastPongTime             time.Time
	reconnectEnabled         bool           
	reconnectInterval        time.Duration  
	maxReconnectAttempts     int            
	currentReconnectAttempts int            
}
func NewClient(logger *utils.Logger) *Client {
	return &Client{
		logger:              logger,
		messageCallbacks:    make(map[string][]MessageCallback),
		connectionCallbacks: make([]ConnectionCallback, 0),
		pendingRequests:     make(map[string]*RequestCallback),
		stopCh:              make(chan struct{}),
		doneCh:              make(chan struct{}),
		reconnectEnabled:         true,
		reconnectInterval:        5 * time.Second,  
		maxReconnectAttempts:     10,               
		currentReconnectAttempts: 0,
	}
}
func (c *Client) Connect(host string, port int, token string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.connected {
		return fmt.Errorf("already connected")
	}
	c.host = host
	c.port = port
	c.token = token
	c.currentReconnectAttempts = 0  
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/plugin_GUI",
	}
	query := u.Query()
	query.Add("token", token)
	u.RawQuery = query.Encode()
	c.logger.Info("正在连接到: %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	c.conn = conn
	c.connected = true
	c.lastPongTime = time.Now()
	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})
	go c.messageLoop()
	go c.heartbeatLoop()
	c.logger.Info("连接成功")
	c.notifyConnectionChange(true)
	return nil
}
func (c *Client) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.connected {
		return
	}
	close(c.stopCh)
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.logger.Info("连接已断开")
	c.notifyConnectionChange(false)
	c.requestMutex.Lock()
	for msgID, callback := range c.pendingRequests {
		if callback.Error != nil {
			callback.Error(fmt.Errorf("client is shutting down"))
		}
		delete(c.pendingRequests, msgID)
	}
	c.requestMutex.Unlock()
	<-c.doneCh
}
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}
func (c *Client) SendRequest(method string, params map[string]interface{}) error {
	return c.SendRequestWithCallback(method, params, nil)
}
func (c *Client) SendRequestWithCallback(method string, params map[string]interface{}, callback *RequestCallback) error {
	return c.SendRequestWithCallbackTimeout(method, params, callback, 3*time.Second)
}
func (c *Client) SendRequestWithCallbackTimeout(method string, params map[string]interface{}, callback *RequestCallback, timeout time.Duration) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	msgID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	header := NewRequestHeader(msgID, nil)
	payload := RequestPayload{
		Method: method,
		Params: params,
	}
	if callback != nil {
		c.requestMutex.Lock()
		c.pendingRequests[msgID] = callback
		c.requestMutex.Unlock()
		go func() {
			time.Sleep(timeout)
			c.requestMutex.Lock()
			if cb, exists := c.pendingRequests[msgID]; exists {
				delete(c.pendingRequests, msgID)
				c.requestMutex.Unlock()
				if cb.Error != nil {
					cb.Error(fmt.Errorf("request timeout for %s", msgID))
				}
			} else {
				c.requestMutex.Unlock()
			}
		}()
	}
	c.logger.Debug("Sending request: %s", method)
	return c.sendMessage(header, payload)
}
func (c *Client) SendResponse(correlationID string, code int, data interface{}, errMsg *string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	msgID := fmt.Sprintf("resp-%d", time.Now().UnixNano())
	header := NewResponseHeader(msgID, correlationID)
	payload := ResponsePayload{
		Code:  code,
		Time:  time.Now().UnixMilli(),
		Data:  data,
		Error: errMsg,
	}
	return c.sendMessage(header, payload)
}
func (c *Client) sendMessage(header MessageHeader, payload interface{}) error {
	c.mutex.RLock()
	conn := c.conn
	c.mutex.RUnlock()
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}
	message, err := EncodeMessage(header, payload)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	c.logger.Debug("发送消息: %s", header.MsgType)
	return nil
}
func (c *Client) messageLoop() {
	defer func() {
		close(c.doneCh)
	}()
	for {
		select {
		case <-c.stopCh:
			return
		default:
			c.mutex.RLock()
			conn := c.conn
			c.mutex.RUnlock()
			if conn == nil {
				return
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					c.logger.Info("连接正常关闭")
				} else {
					c.logger.Error("读取消息失败: %v", err)
				}
				c.handleDisconnection()
				return
			}
			c.handleMessage(string(message))
		}
	}
}
func (c *Client) handleMessage(rawData string) {
	header, payload, err := DecodeMessage(rawData)
	if err != nil {
		c.logger.Error("解码消息失败: %v", err)
		return
	}
	c.logger.Debug("收到消息: %s", header.MsgType)
	if header.MsgType == "response" && header.CorrelationID != nil {
		c.handleResponse(*header.CorrelationID, payload)
		return
	}
	c.mutex.RLock()
	callbacks := c.messageCallbacks[header.MsgType]
	c.mutex.RUnlock()
	for _, callback := range callbacks {
		go callback(*header, payload)
	}
}
func (c *Client) handleResponse(msgID string, payload interface{}) {
	c.requestMutex.Lock()
	callback, exists := c.pendingRequests[msgID]
	if exists {
		delete(c.pendingRequests, msgID)
	}
	c.requestMutex.Unlock()
	if exists && callback != nil {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			if errMsg, hasError := payloadMap["error"]; hasError && errMsg != nil && errMsg != "<nil>" {
				if callback.Error != nil {
					callback.Error(fmt.Errorf("server error: %v", errMsg))
				}
			} else if callback.Success != nil {
				callback.Success(payload)
			}
		} else if callback.Success != nil {
			callback.Success(payload)
		}
	}
}
func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if !c.IsConnected() {
				return
			}
			msgID := fmt.Sprintf("hb-%d", time.Now().UnixNano())
			header := MessageHeader{
				MsgID:     msgID,
				MsgType:   "heartbeat",
				Timestamp: float64(time.Now().UnixNano()) / 1e9,
			}
			payload := map[string]interface{}{"status": "alive"}
			if err := c.sendMessage(header, payload); err != nil {
				c.logger.Error("发送心跳失败: %v", err)
				c.handleDisconnection()
				return
			}
		}
	}
}
func (c *Client) handleDisconnection() {
	c.mutex.Lock()
	wasConnected := c.connected
	if c.connected {
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.logger.Info("连接已断开")
		c.notifyConnectionChange(false)
	}
	c.mutex.Unlock()
	c.requestMutex.Lock()
	for msgID, callback := range c.pendingRequests {
		if callback.Error != nil {
			callback.Error(fmt.Errorf("connection lost"))
		}
		delete(c.pendingRequests, msgID)
	}
	c.requestMutex.Unlock()
	if wasConnected && c.reconnectEnabled && c.currentReconnectAttempts < c.maxReconnectAttempts {
		c.currentReconnectAttempts++
		c.logger.Info("尝试自动重连 (%d/%d)...", c.currentReconnectAttempts, c.maxReconnectAttempts)
		go func() {
			time.Sleep(c.reconnectInterval)
			if err := c.Connect(c.host, c.port, c.token); err != nil {
				c.logger.Error("自动重连失败: %v", err)
			}
		}()
	}
}
func (c *Client) OnConnectionChanged(callback ConnectionCallback) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connectionCallbacks = append(c.connectionCallbacks, callback)
}
func (c *Client) OnMessage(msgType string, callback MessageCallback) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messageCallbacks[msgType] = append(c.messageCallbacks[msgType], callback)
}
func (c *Client) notifyConnectionChange(connected bool) {
	for _, callback := range c.connectionCallbacks {
		go callback(connected)
	}
}
func (c *Client) SetReconnectEnabled(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.reconnectEnabled = enabled
}
func (c *Client) SetReconnectParams(interval time.Duration, maxAttempts int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.reconnectInterval = interval
	c.maxReconnectAttempts = maxAttempts
}
func (c *Client) GetReconnectStatus() (bool, int, int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.reconnectEnabled, c.currentReconnectAttempts, c.maxReconnectAttempts
}