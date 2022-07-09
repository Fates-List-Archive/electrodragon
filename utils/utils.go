package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/alexedwards/argon2id"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"github.com/pquerna/otp/totp"

	"golang.org/x/exp/constraints"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Schema struct {
	TableName  string  `json:"table_name"`
	ColumnName string  `json:"column_name"`
	Type       string  `json:"type"`
	IsNullable bool    `json:"nullable"`
	Array      bool    `json:"array"`
	DefaultSQL *string `json:"default_sql"`
	DefaultVal any     `json:"default_val"`
	Secret     bool    `json:"secret"`
}

func IsSecret(tableName, columnName string) bool {
	colArray := [2]string{tableName, columnName}

	secretCols := [][2]string{
		{
			"bots", "api_token",
		},
		{
			"bots", "webhook_secret",
		},
		{
			"users", "api_token",
		},
		{
			"users", "staff_password",
		},
		{
			"users", "totp_shared_key",
		},
		{
			"users", "supabase_id",
		},
		{
			"servers", "api_token",
		},
		{
			"servers", "webhook_secret",
		},
	}

	for _, col := range secretCols {
		if colArray == col {
			return true
		}
	}
	return false
}

type schemaData struct {
	ColumnDefault *string `db:"column_default"`
	TableName     string  `db:"table_name"`
	ColumnName    string  `db:"column_name"`
	DataType      string  `db:"data_type"`
	ElementType   *string `db:"element_type"`
	IsNullable    string  `db:"is_nullable"`
}

// Filter the postgres schema
type SchemaFilter struct {
	TableName string
}

func GetSchema(ctx context.Context, pool *pgxpool.Pool, opts SchemaFilter) ([]Schema, error) {
	var sqlString string = `
	SELECT c.is_nullable, c.table_name, c.column_name, c.column_default, c.data_type AS data_type, e.data_type AS element_type FROM information_schema.columns c LEFT JOIN information_schema.element_types e
	ON ((c.table_catalog, c.table_schema, c.table_name, 'TABLE', c.dtd_identifier)
= (e.object_catalog, e.object_schema, e.object_name, e.object_type, e.collection_type_identifier))
WHERE table_schema = 'public' order by table_name, ordinal_position
`
	rows, err := pool.Query(ctx, sqlString)

	if err != nil {
		return nil, err
	}

	var result []Schema

	for rows.Next() {
		var schema Schema

		data := schemaData{}

		err := rows.Scan(&data.IsNullable, &data.TableName, &data.ColumnName, &data.ColumnDefault, &data.DataType, &data.ElementType)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if opts.TableName != "" && opts.TableName != data.TableName {
			fmt.Println("Ignoring table", data.TableName, "as it is not", opts.TableName)
			continue
		}

		// Create new transaction to get default column
		if data.ColumnDefault != nil && *data.ColumnDefault != "" {
			tx, err := pool.Begin(ctx)
			if err != nil {
				return nil, err
			}

			var defaultV any

			err = tx.QueryRow(ctx, "SELECT "+*data.ColumnDefault).Scan(&defaultV)

			if err != nil {
				return nil, err
			}

			fmt.Println(data.ColumnName, reflect.TypeOf(defaultV))

			err = tx.Rollback(ctx)

			if err != nil {
				return nil, err
			}

			// Check for [16]uint8 case
			if defaultVal, ok := defaultV.([16]uint8); ok {
				defaultV = fmt.Sprintf("%x-%x-%x-%x-%x", defaultVal[0:4], defaultVal[4:6], defaultVal[6:8], defaultVal[8:10], defaultVal[10:16])
			}

			schema.DefaultVal = defaultV
		} else {
			schema.DefaultVal = nil
		}

		// Now check if the column is tagged properly
		var tag pgtype.UUID
		if err := pool.QueryRow(ctx, "SELECT _lynxtag FROM"+data.TableName).Scan(&tag); err != nil {
			if err == pgx.ErrNoRows {
				fmt.Println("Tagging", data.TableName)
				_, err := pool.Exec(ctx, "ALTER TABLE "+data.TableName+" ADD COLUMN _lynxtag uuid not null unique default uuid_generate_v4()")
				if err != nil {
					return nil, err
				}
			}
		}

		schema.ColumnName = data.ColumnName
		schema.TableName = data.TableName
		schema.DefaultSQL = data.ColumnDefault

		schema.IsNullable = (data.IsNullable == "YES")

		if data.DataType == "ARRAY" {
			schema.Type = *data.ElementType
			schema.Array = true
		} else {
			schema.Type = data.DataType
		}

		schema.Secret = IsSecret(data.TableName, data.ColumnName)

		result = append(result, schema)
	}

	fmt.Println("Got", len(result), "items")

	return result, nil
}

type UserPerms struct {
	Perm    float64 `json:"perm"`
	ID      string  `json:"id"`
	StaffID string  `json:"staff_id"`
	Fname   string  `json:"fname"`
}

// Gets the permissions of a user from baypaw
func GetPermissions(devMode bool, userID string) (*UserPerms, error) {
	var api string
	if devMode {
		api = "https://api.fateslist.xyz/baypaw/perms/"
	} else {
		api = "http://localhost:1234/perms/"
	}
	req, err := http.NewRequest("GET", api+userID, nil)

	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var perms UserPerms

	err = json.NewDecoder(resp.Body).Decode(&perms)

	if err != nil {
		return nil, err
	}

	return &perms, nil
}

type sessionStruct struct {
	ID    string `json:"user_id"`
	Token string `json:"token"`
}

type AuthRequest struct {
	// The users ID
	UserID string

	// User API token
	Token string

	// Session ID (after login for some extra secure endpoints)
	SessionID string

	// User Password
	Password string

	// 2FA code
	TOTP string

	// Developer mode or not
	DevMode bool

	// The context of the request
	Context context.Context

	// The pgx database used
	DB *pgxpool.Pool

	// Redis client (requied for session ID)
	Redis *redis.Client
}

type authResponse struct {
	// The users permissions
	Perms UserPerms

	// If the user is staff verified, this will be set to true
	Verified bool

	// If the user is MFA key verified or not
	MFA bool

	// If the user has logged in with a password successfully
	PasswordLogin bool

	// The allowed tables of the user, empty slice if all are allowed
	AllowedTables []string

	// Whether or not the users session was validated or not
	SessionValidated bool
}

// Helper function to authenticate a user
func AuthorizeUser(req AuthRequest) (*authResponse, error) {
	if req.Token == "" {
		return nil, errors.New("no token provided")
	}

	perms, err := GetPermissions(req.DevMode, req.UserID)

	if err != nil {
		return nil, err
	}

	// Check auth token
	var count int
	err = req.DB.QueryRow(req.Context, "SELECT COUNT(1) FROM users WHERE user_id = $1 AND api_token = $2", req.UserID, strings.ReplaceAll(req.Token, " ", "")).Scan(&count)

	if err != nil {
		return nil, err
	}

	if count > 1 {
		// Delete all other users with this ID
		_, err = req.DB.Exec(req.Context, "DELETE FROM users WHERE user_id = $1", req.UserID)

		if err != nil {
			return nil, err
		}

		return nil, errors.New("multiple users with this ID. Retry logging in now as all users with this ID have been deleted")
	}

	if count == 0 {
		return nil, errors.New("invalid token")
	}

	// Check staff verify code
	var staffVerifyCode string

	req.DB.QueryRow(req.Context, "SELECT staff_verify_code FROM users WHERE user_id = $1", req.UserID).Scan(&staffVerifyCode)

	var verified bool

	if req.DevMode {
		verified = CheckCodeDev(req.UserID, staffVerifyCode)
	} else {
		verified = CheckCodeSecure(req.UserID, staffVerifyCode)
	}

	// Check MFA
	var totpKey string
	var mfa bool

	req.DB.QueryRow(req.Context, "SELECT totp_shared_key FROM users WHERE user_id = $1", req.UserID).Scan(&totpKey)

	if totpKey != "" && req.TOTP != "" && totp.Validate(req.TOTP, totpKey) {
		mfa = true
	}

	fmt.Println(totpKey, req.TOTP, mfa)

	// Check password
	var passAuth bool

	if req.Password != "" {
		// Get argon2 password from DB

		var password string

		req.DB.QueryRow(req.Context, "SELECT staff_password FROM users WHERE user_id = $1", req.UserID).Scan(&password)

		if password != "" {
			if match, err := argon2id.ComparePasswordAndHash(req.Password, password); err == nil && match {
				passAuth = true
			}
		}
	}

	// SessionValidated is true if the session was validated
	var sessionValidated bool

	if req.SessionID != "" {
		// Check session in redis
		var session = req.Redis.Get(req.Context, req.SessionID).Val()

		if session != "" {
			// JSON parse it (or try to)
			var sessionData sessionStruct

			err := json.Unmarshal([]byte(session), &sessionData)

			if err != nil {
				fmt.Println(err)
			} else {
				if sessionData.ID != req.UserID {
					return nil, errors.New("invalid session")
				} else {
					// Check token with the token that was in auth request (that was also validated)
					if sessionData.Token != req.Token {
						return nil, errors.New("invalid session")
					}
					sessionValidated = true
				}
			}
		}
	}

	allowedTokens := []string{}

	if perms.Perm < 5 {
		allowedTokens = []string{"reviews", "review_votes", "bot_packs", "vanity", "leave_of_absence", "user_vote_table",
			"lynx_surveys", "lynx_survey_responses"}
	}

	resp := &authResponse{
		Perms:            *perms,
		Verified:         verified,
		MFA:              mfa,
		AllowedTables:    allowedTokens,
		PasswordLogin:    passAuth,
		SessionValidated: sessionValidated,
	}

	return resp, nil
}

type HFunc = func(w http.ResponseWriter, r *http.Request)

func CorsWrap(fn HFunc) HFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Frostpaw-ID, Frostpaw-MFA, Authorization, Frostpaw-Pass")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.Write([]byte(""))
			return
		}
		fn(w, r)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
