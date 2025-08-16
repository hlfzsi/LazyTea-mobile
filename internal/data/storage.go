package data
import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)
type ConnectionConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Token    string `json:"token"`
	Remember bool   `json:"remember"`
}
type BotInfo struct {
	ID          string    `json:"id"`
	AdapterName string    `json:"adapter_name"`
	Platform    string    `json:"platform"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
	DisplayName string    `json:"display_name"`
	Avatar      string    `json:"avatar"`
}
type Message struct {
	ID         int64   `json:"id"`                  
	User       string  `json:"user"`                
	GroupID    *string `json:"group_id,omitempty"`  
	Bot        string  `json:"bot"`                 
	Timestamps int64   `json:"timestamps"`          
	Content    string  `json:"content"`             
	Meta       *string `json:"meta,omitempty"`      
	Plaintext  string  `json:"plaintext"`           
	BotID     string    `json:"bot_id"`                
	FromBot   bool      `json:"from_bot"`              
	Timestamp time.Time `json:"timestamp"`             
	UserID    string    `json:"user_id"`               
	UserName  string    `json:"user_name"`             
	GroupName *string   `json:"group_name,omitempty"`  
}
type Plugin struct {
	Name        string      `json:"name"`
	Module      string      `json:"module"`
	Meta        PluginMeta  `json:"meta"`
}
type PluginMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	ConfigExist bool   `json:"config_exist"`
	IconAbspath string `json:"icon_abspath"`
	Author      string `json:"author"`
	Version     string `json:"version"`
}
type Storage struct {
	db    *sql.DB
	mutex sync.RWMutex  
	ftsEnabled bool
	botInfos map[string]BotInfo
}
type PluginCallRecord struct {
	Bot             string
	Platform        string
	TimeCosted      float64
	GroupID         *string
	UserID          *string
	PluginName      string
	MatcherHash     string
	ExceptionName   *string
	ExceptionDetail *string
	Timestamp       int64
}
func NewStorage(dbPath string) (*Storage, error) {
	dbPath += "?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=1"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(25)                  
	db.SetMaxIdleConns(5)                   
	db.SetConnMaxLifetime(5 * time.Minute)  
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	storage := &Storage{
		db:       db,
		botInfos: make(map[string]BotInfo),
	}
	if err := storage.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}
	return storage, nil
}
func (s *Storage) Close() error {
	return s.db.Close()
}
func (s *Storage) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS connection_config (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            host TEXT NOT NULL,
            port INTEGER NOT NULL,
            token TEXT NOT NULL,
            remember BOOLEAN DEFAULT FALSE,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS Message (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user TEXT,
            group_id TEXT,
            bot TEXT,
            timestamps INTEGER,
            content TEXT,
            meta TEXT,
            plaintext TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS plugin_call_record (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            bot TEXT,
            platform TEXT,
            time_costed REAL,
            group_id TEXT,
            user_id TEXT,
            plugin_name TEXT,
            matcher_hash TEXT,
            exception_name TEXT,
            exception_detail TEXT,
            timestamp INTEGER
        )`,
		`CREATE INDEX IF NOT EXISTS idx_bot_platform ON plugin_call_record (bot, platform)`,
		`CREATE INDEX IF NOT EXISTS idx_timestamp ON plugin_call_record (timestamp)`,
	}
	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	if err := s.enableFTS5(); err != nil {
		if err := s.enablePlainIndex(); err != nil {
			return err
		}
		s.ftsEnabled = false
	} else {
		s.ftsEnabled = true
	}
	_, _ = s.db.Exec("REINDEX;")
	return nil
}
func (s *Storage) enableFTS5() error {
	stmts := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS message_for_fts USING fts5(plaintext)`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_insert AFTER INSERT ON Message
         BEGIN
             INSERT INTO message_for_fts(rowid, plaintext)
             VALUES (NEW.id, NEW.plaintext);
         END;`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_update AFTER UPDATE ON Message
         BEGIN
             UPDATE message_for_fts
             SET plaintext = NEW.plaintext
             WHERE rowid = NEW.id;
         END;`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_delete AFTER DELETE ON Message
         BEGIN
             DELETE FROM message_for_fts WHERE rowid = OLD.id;
         END;`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
func (s *Storage) enablePlainIndex() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS message_for_fts (
            rowid INTEGER PRIMARY KEY,
            plaintext TEXT
        )`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_insert AFTER INSERT ON Message
         BEGIN
             INSERT OR REPLACE INTO message_for_fts(rowid, plaintext)
             VALUES (NEW.id, NEW.plaintext);
         END;`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_update AFTER UPDATE ON Message
         BEGIN
             UPDATE message_for_fts
             SET plaintext = NEW.plaintext
             WHERE rowid = NEW.id;
         END;`,
		`CREATE TRIGGER IF NOT EXISTS trigger_message_delete AFTER DELETE ON Message
         BEGIN
             DELETE FROM message_for_fts WHERE rowid = OLD.id;
         END;`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to enable plain index: %w", err)
		}
	}
	return nil
}
func (s *Storage) SaveConnectionConfig(config ConnectionConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, err := s.db.Exec("DELETE FROM connection_config"); err != nil {
		return fmt.Errorf("failed to delete old config: %w", err)
	}
	query := `INSERT INTO connection_config (host, port, token, remember) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, config.Host, config.Port, config.Token, config.Remember)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
func (s *Storage) LoadConnectionConfig() (*ConnectionConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	query := `SELECT host, port, token, remember FROM connection_config ORDER BY updated_at DESC LIMIT 1`
	var config ConnectionConfig
	err := s.db.QueryRow(query).Scan(&config.Host, &config.Port, &config.Token, &config.Remember)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no config found")
		}
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &config, nil
}
func (s *Storage) SaveBotInfo(bot BotInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.botInfos[bot.ID] = bot
	return nil
}
func (s *Storage) GetBotInfoList() ([]BotInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	result := make([]BotInfo, 0, len(s.botInfos))
	for _, bot := range s.botInfos {
		result = append(result, bot)
	}
	return result, nil
}
func (s *Storage) UpdateBotOnlineStatus(botID string, isOnline bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if bot, exists := s.botInfos[botID]; exists {
		bot.IsOnline = isOnline
		bot.LastSeen = time.Now()
		s.botInfos[botID] = bot
	}
	return nil
}
func (s *Storage) SaveMessage(msg Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	query := `INSERT INTO Message 
        (user, group_id, bot, timestamps, content, meta, plaintext) 
        VALUES (?, ?, ?, ?, ?, ?, ?)`
	timestamps := msg.Timestamps
	if timestamps == 0 {
		timestamps = time.Now().UnixMilli()
	}
	bot := msg.Bot
	if bot == "" {
		bot = msg.BotID
	}
	user := msg.User
	if user == "" {
		user = msg.UserID
	}
	plaintext := msg.Plaintext
	if plaintext == "" {
		plaintext = msg.Content
	}
	_, err := s.db.Exec(query, user, msg.GroupID, bot, timestamps,
		msg.Content, msg.Meta, plaintext)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}
func (s *Storage) GetMessages(limit int, offset int) ([]Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	query := `SELECT id, COALESCE(user, ''), COALESCE(group_id, ''), bot, 
        timestamps, content, COALESCE(meta, ''), COALESCE(plaintext, content) 
        FROM Message ORDER BY id ASC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		var groupID string
		var meta string
		err := rows.Scan(&msg.ID, &msg.User, &groupID, &msg.Bot,
			&msg.Timestamps, &msg.Content, &meta, &msg.Plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.BotID = msg.Bot
		msg.UserID = msg.User
		msg.Timestamp = time.UnixMilli(msg.Timestamps)
		if groupID != "" {
			msg.GroupID = &groupID
		}
		if meta != "" {
			msg.Meta = &meta
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
func (s *Storage) SearchMessages(query string, limit int) ([]Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.ftsEnabled {
		isFTS5, err := s.detectFTSType()
		if err != nil {
			return nil, fmt.Errorf("failed to detect FTS type: %w", err)
		}
		if isFTS5 {
			return s.searchWithFTS5(query, limit)
		} else {
			return s.searchWithLikeQuery(query, limit)
		}
	}
	return s.searchWithLikeQuery(query, limit)
}
func (s *Storage) tokenizeSearchQuery(query string) (string, error) {
	tokenizedStr, err := goTokenizeForFts(query)
	if err != nil {
		return "", fmt.Errorf("failed to tokenize query: %w", err)
	}
	if tokenizedStr == "" {
		return "", nil
	}
	tokens := strings.Fields(tokenizedStr)
	if len(tokens) == 0 {
		return "", nil
	}
	if len(tokens) == 0 {
		return "", nil
	}
	finalQuery := strings.Join(tokens, " OR ")
	return finalQuery, nil
}
func (s *Storage) IsFTSEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.ftsEnabled
}
func (s *Storage) detectFTSType() (bool, error) {
	var sql_text string
	err := s.db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='message_for_fts'").Scan(&sql_text)
	if err != nil {
		return false, fmt.Errorf("failed to get table definition: %w", err)
	}
	isFTS5 := strings.Contains(strings.ToLower(sql_text), "fts5")
	return isFTS5, nil
}
func (s *Storage) searchWithFTS5(query string, limit int) ([]Message, error) {
	tokenizedQuery, err := s.tokenizeSearchQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize search query: %w", err)
	}
	if tokenizedQuery == "" {
		return []Message{}, nil
	}
	sqlQuery := `
        SELECT 
            m.id,
            COALESCE(m.user, ''),
            COALESCE(m.group_id, ''),
            m.bot,
            m.timestamps,
            m.content,
            COALESCE(m.meta, ''),
            COALESCE(m.plaintext, m.content)
        FROM
            message_for_fts f
        JOIN
            Message m ON f.rowid = m.id
        WHERE
            f.plaintext MATCH ?
        ORDER BY
            m.timestamps DESC
        LIMIT ?`
	rows, err := s.db.Query(sqlQuery, tokenizedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages with FTS5: %w", err)
	}
	defer rows.Close()
	return s.scanMessageRows(rows)
}
func (s *Storage) searchWithLikeQuery(query string, limit int) ([]Message, error) {
	tokenizedStr, err := goTokenizeForFts(query)
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize query: %w", err)
	}
	tokens := strings.Fields(tokenizedStr)
	if len(tokens) == 0 {
		return []Message{}, nil
	}
	var whereConditions []string
	var args []interface{}
	for _, token := range tokens {
		whereConditions = append(whereConditions, "f.plaintext LIKE ?")
		args = append(args, "%"+token+"%")
	}
	args = append(args, limit)
	sqlQuery := `
        SELECT 
            m.id,
            COALESCE(m.user, ''),
            COALESCE(m.group_id, ''),
            m.bot,
            m.timestamps,
            m.content,
            COALESCE(m.meta, ''),
            COALESCE(m.plaintext, m.content)
        FROM
            message_for_fts f
        JOIN
            Message m ON f.rowid = m.id
        WHERE ` + strings.Join(whereConditions, " AND ") + `
        ORDER BY
            m.timestamps DESC
        LIMIT ?`
	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages with LIKE: %w", err)
	}
	defer rows.Close()
	return s.scanMessageRows(rows)
}
func (s *Storage) scanMessageRows(rows *sql.Rows) ([]Message, error) {
	var messages []Message
	rowCount := 0
	for rows.Next() {
		rowCount++
		var msg Message
		var groupID string
		var meta string
		err := rows.Scan(&msg.ID, &msg.User, &groupID, &msg.Bot, &msg.Timestamps, &msg.Content, &meta, &msg.Plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.BotID = msg.Bot
		msg.UserID = msg.User
		msg.Timestamp = time.UnixMilli(msg.Timestamps)
		if groupID != "" {
			msg.GroupID = &groupID
		}
		if meta != "" {
			msg.Meta = &meta
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
func (s *Storage) GetMessagesByBot(botID string, limit int, offset int) ([]Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	query := `SELECT id, COALESCE(user, ''), COALESCE(group_id, ''), bot, 
        timestamps, content, COALESCE(meta, ''), COALESCE(plaintext, content) 
        FROM Message WHERE bot = ? 
        ORDER BY timestamps DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, botID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by bot: %w", err)
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		var groupID string
		var meta string
		err := rows.Scan(&msg.ID, &msg.User, &groupID, &msg.Bot,
			&msg.Timestamps, &msg.Content, &meta, &msg.Plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.BotID = msg.Bot
		msg.UserID = msg.User
		msg.Timestamp = time.UnixMilli(msg.Timestamps)
		if groupID != "" {
			msg.GroupID = &groupID
		}
		if meta != "" {
			msg.Meta = &meta
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
func (s *Storage) GetBotMessageCount(botID string) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var count int64
	query := `SELECT COUNT(*) FROM Message WHERE bot = ?`
	err := s.db.QueryRow(query, botID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get bot message count: %w", err)
	}
	return count, nil
}
func (s *Storage) GetBotMessageCountInPeriod(botID string, minutes int) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var count int64
	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute).UnixMilli()
	query := `SELECT COUNT(*) FROM Message WHERE bot = ? AND timestamps >= ?`
	err := s.db.QueryRow(query, botID, cutoffTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get bot message count in period: %w", err)
	}
	return count, nil
}
func (s *Storage) GetTotalMessageCount() (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var count int64
	query := `SELECT COUNT(*) FROM Message`
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total message count: %w", err)
	}
	return count, nil
}
func (s *Storage) SavePluginCall(rec PluginCallRecord) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	query := `INSERT INTO plugin_call_record(
        bot, platform, time_costed, group_id, user_id, plugin_name,
        matcher_hash, exception_name, exception_detail, timestamp
    ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query,
		rec.Bot, rec.Platform, rec.TimeCosted, rec.GroupID, rec.UserID, rec.PluginName,
		rec.MatcherHash, rec.ExceptionName, rec.ExceptionDetail, rec.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save plugin_call_record: %w", err)
	}
	return nil
}
func (s *Storage) SavePlugin(plugin Plugin) error { return nil }
func (s *Storage) GetPluginList() ([]Plugin, error) { return []Plugin{}, nil }
func (s *Storage) ClearAllData() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	tables := []string{
		"Message",
		"message_for_fts",
		"plugin_call_record", 
		"connection_config",
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := tx.Exec(query); err != nil {
			if !strings.Contains(err.Error(), "no such table") {
				return fmt.Errorf("failed to clear table %s: %w", table, err)
			}
		}
	}
	resetQuery := "DELETE FROM sqlite_sequence WHERE name IN ('Message', 'plugin_call_record', 'connection_config')"
	if _, err := tx.Exec(resetQuery); err != nil {
		if !strings.Contains(err.Error(), "no such table") {
			return fmt.Errorf("failed to reset sequence: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	if _, err := s.db.Exec("VACUUM"); err != nil {
	}
	return nil
}
