package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/elliotchance/orderedmap"
	"github.com/labstack/gommon/log"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/yanzay/tbot/v2"
)

func (a *application) setCacheItem(key string, val interface{}) {
	a.cache.Set(key, val)
}

func (a *application) getCacheItem(key string) interface{} {
	val, _ := a.cache.Get(key)
	return val
}

func (a *application) delCacheItem(key string) {
	a.cache.Remove(key)
}

func (a *application) getFileLink(fileID string) string {
	file, err := a.tgClient.GetFileInfo(fileID)
	if err != nil {
		log.Error("Error getting file link ", err)
		return ""
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", a.token, file.FilePath)
}

func (a *application) messageHandler(m *tbot.Message) {
	a.handleMessage(m)
}

func (a *application) callbackHandler(cq *tbot.CallbackQuery) {
	a.tgClient.AnswerCallback(cq.ID)

	a.handleCallback(cq)
}

func (a *application) replaceInlineOptions(chatID string, msgID int, inlineOptions []map[string]interface{}) int {
	id, err := a.tgClient.EditInlineMarkup(chatID, msgID, buildInlineOptions(inlineOptions))
	if err != nil {
		log.Error("Error replacing inline options ", err)
	}
	return id
}

func (a *application) deleteMessage(chatID string, msgID int) {
	err := a.tgClient.DeleteMsg(chatID, msgID)
	if err != nil {
		log.Error("Error deleting message ", err)
	}
}

func (a *application) editMessage(chatID string, msgID int, text string, inlineOptions []map[string]interface{}) {
	err := a.tgClient.EditMsg(chatID, msgID, text, buildInlineOptions(inlineOptions))
	if err != nil {
		log.Error("Error editing message ", err)
	}
}

func (a *application) doGet(aURL string, params map[string]interface{}, headers map[string]interface{}, timeoutSec int) string {
	resp, err := doGET(aURL, params, headers, timeoutSec)
	if err != nil {
		log.Error("Error performing GET request ", err)
	}

	return resp
}

func (a *application) doPOST(aURL string, params map[string]interface{}, headers map[string]interface{}, timeoutSec int) string {
	resp, err := doPOST(aURL, params, headers, timeoutSec)
	if err != nil {
		log.Error("Error performing POST request ", err)
	}

	return resp
}

func (a *application) ReportDB(userID string, text string, query string, name string, args []interface{}) int {
	results := a.QueryDB(query, args)

	resLen := len(results)
	if resLen > 0 {
		report := make([][]string, resLen+1)
		for id, row := range results {
			if id == 0 {
				//append column names
				report[0] = make([]string, 0, row.Len())
				for el := row.Front(); el != nil; el = el.Next() {
					report[0] = append(report[0], fmt.Sprintf("%s", el.Key))
				}
			}
			report[id+1] = make([]string, 0, row.Len())
			//append values
			for _, key := range report[0] {
				report[id+1] = append(report[id+1], fmt.Sprintf("%v", row.GetOrDefault(key, "")))
			}
		}
		file, err := ioutil.TempFile(a.attachmentsDir, fmt.Sprintf("%s*.csv", name))
		if err != nil {
			log.Error("Error creating report on disk ", err)
		} else {
			defer os.Remove(file.Name())
			defer file.Close()
			w := csv.NewWriter(file)
			if err := w.WriteAll(report); err != nil {
				log.Error("Error saving report to disk ", err)
			} else {
				_, basename := filepath.Split(file.Name())
				return a.sendMessage(userID, text, [][]string{}, []map[string]interface{}{}, basename)
			}
		}

	}

	log.Info("Skipped empty resultset in report generation")
	return 0
}

func (a *application) QueryDB(query string, args []interface{}) []*orderedmap.OrderedMap {
	result := []*orderedmap.OrderedMap{}
	if a.dbClient != nil {
		rows, err := a.dbClient.Query(query, args...)
		if err != nil {
			log.Error("Error querying db ", err)
			return result
		}
		defer rows.Close()

		columnTypes, err := rows.ColumnTypes()

		if err != nil {
			log.Error("Error querying db ", err)
			return result
		}

		count := len(columnTypes)

		for rows.Next() {

			scanArgs := make([]interface{}, count)

			for i, v := range columnTypes {

				switch v.DatabaseTypeName() {
				case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
					scanArgs[i] = new(sql.NullString)
					break
				case "BOOL":
					scanArgs[i] = new(sql.NullBool)
					break
				case "INT4":
					scanArgs[i] = new(sql.NullInt64)
					break
				default:
					scanArgs[i] = new(sql.NullString)
				}
			}

			err := rows.Scan(scanArgs...)

			if err != nil {
				log.Error("Error querying db ", err)
				return result
			}

			masterData := orderedmap.NewOrderedMap()

			for i, v := range columnTypes {

				if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
					masterData.Set(v.Name(), z.Bool)
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					masterData.Set(v.Name(), z.String)
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
					masterData.Set(v.Name(), z.Int64)
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
					masterData.Set(v.Name(), z.Float64)
					continue
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
					masterData.Set(v.Name(), z.Int32)
					continue
				}

				masterData.Set(v.Name(), scanArgs[i])
			}

			result = append(result, masterData)
		}
	}
	return result
}

func (a *application) ExecDB(query string, args []interface{}) sql.Result {
	if a.dbClient != nil {
		res, err := a.dbClient.Exec(query, args...)
		if err != nil {
			log.Error("Error executing db query ", err)
		} else {
			return res
		}
	}

	return nil
}

func (a *application) initialize() error {

	//prepare js runtime
	if GetEnv("SCRIPTS", "") == "" {
		return errors.New("No scripts are configured")
	}

	scripts := strings.Split(GetEnv("SCRIPTS", ""), ",")

	var b bytes.Buffer
	for _, scriptPath := range scripts {
		script, err := ReadFile(scriptPath)
		if err != nil {
			return err
		}
		b.WriteString(script)
		b.WriteString("\n")
	}

	a.vmTemplate = a.createVmTemplate()
	if _, err := a.vmTemplate.Run(b.String()); err != nil {
		return err
	}
	if _, err := a.vmTemplate.Object("bot"); err != nil {
		return err
	}

	//setup DB connection
	if GetEnv("DB_DRIVER", "") != "" && GetEnv("DB_CONN_STR", "") != "" {
		var err error
		if a.dbClient, err = sql.Open(GetEnv("DB_DRIVER", ""), GetEnv("DB_CONN_STR", "")); err != nil {
			return err
		}
		if err = a.dbClient.Ping(); err != nil {
			return err
		}
	}

	//configure cache
	a.cache = ttlcache.NewCache()
	duration, err := time.ParseDuration(GetEnv("CACHE_TTL", "30m"))
	if err != nil {
		log.Error("Error parsing time duration for cache ttl, ttl set to 30 minutes ", err)
		duration = time.Minute * 30
	}
	a.cache.SetTTL(duration)

	return nil
}

func (a *application) GetBot(id string) *otto.Object {
	vm := a.vmTemplate.Copy()

	if id != "" {
		vm.Set("send", a.getSendFunc(id))

		vm.Set("prompt", a.getPromptFunc(id))

		vm.Set("set", a.getSetFunc(id))

		vm.Set("get", a.getGetFunc(id))

		vm.Set("del", a.getDelFunc(id))

		vm.Set("dbReport", a.getReportDBFunc(id))
	}

	bot, _ := vm.Object("bot")

	return bot
}

func (a *application) createVmTemplate() Vm {

	vm := a.vmFactory.GetVm()

	vm.Set("doGet", a.getDoGetFunc())

	vm.Set("doPost", a.getDoPostFunc())

	vm.Set("dbQuery", a.getQueryDBFunc())

	vm.Set("dbExec", a.getExecDBFunc())

	vm.Set("dbReport", a.getReportDBFunc(""))

	vm.Set("getFileLink", a.getGetFileLinkFunc())

	vm.Set("replaceOptions", a.getReplaceOptionsFunc())

	vm.Set("deleteMessage", a.getDeleteMessageFunc())

	vm.Set("editMessage", a.getEditMessageFunc())

	vm.Set("sleep", a.getSleepFunc())

	vm.Set("env", a.getEnvFunc())

	vm.Set("send", a.getSendFunc(""))

	vm.Set("prompt", a.getPromptFunc(""))

	vm.Set("set", a.getSetFunc(""))

	vm.Set("get", a.getGetFunc(""))

	vm.Set("del", a.getDelFunc(""))

	return vm
}

func (a *application) onTimer() {
	_, err := a.GetBot("").Call("onTimer")

	if err != nil {
		log.Error("Error in onTimer ", err)
	}
}

func (a *application) onInit() {
	//start timer here when everything is ready
	if GetEnv("TIMER", "") != "" {
		duration, err := time.ParseDuration(GetEnv("TIMER", ""))
		if err != nil {
			log.Error("Error parsing time duration for timer ", err)
		} else {
			ticker := time.NewTicker(duration)
			go func() {
				for range ticker.C {
					a.onTimer()
				}
			}()
		}
	}

	_, err := a.GetBot("").Call("onInit")

	if err != nil {
		log.Error("Error in onInit ", err)
	}
}

func (a *application) handleMessage(m *tbot.Message) {
	_, err := a.GetBot(m.Chat.ID).Call("onMessage", m)

	if err != nil {
		log.Error("Error in handleMessage ", err)
	}
}

func (a *application) handleCallback(cq *tbot.CallbackQuery) {
	_, err := a.GetBot(cq.Message.Chat.ID).Call("onCallback", cq)

	if err != nil {
		log.Error("Error in handleCallback ", err)
	}
}

func (a *application) getReportDBFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		name := "report"
		if n, err := call.Argument(0).ToString(); err == nil {
			name = n
		}

		text := ""
		if t, err := call.Argument(1).ToString(); err == nil {
			text = t
		}

		targetUser := userID
		if call.Argument(2).IsDefined() && !call.Argument(2).IsNull() {
			if tu, err := call.Argument(2).ToString(); err == nil {
				targetUser = tu
			}
		}

		if query, err := call.Argument(3).ToString(); err == nil {

			var arguments []interface{}
			for i := 4; i < len(call.ArgumentList); i++ {
				arg, _ := call.Argument(i).Export()
				arguments = append(arguments, arg)
			}

			id := a.ReportDB(targetUser, text, query, name, arguments)

			result, _ = otto.ToValue(id)
		}

		return result
	}
}

func (a *application) getQueryDBFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		if query, err := call.Argument(0).ToString(); err == nil {
			var arguments []interface{}
			for i := 1; i < len(call.ArgumentList); i++ {
				arg, _ := call.Argument(i).Export()
				arguments = append(arguments, arg)
			}
			rows := a.QueryDB(query, arguments)
			var out bytes.Buffer
			out.WriteRune('[')
			for idx, row := range rows {
				out.WriteRune('{')
				i := 1
				for el := row.Front(); el != nil; el = el.Next() {
					out.WriteString(strconv.Quote(fmt.Sprintf("%s", el.Key)))
					out.WriteRune(':')
					v, _ := json.Marshal(el.Value)
					out.WriteString(string(v))
					if i < row.Len() {
						out.WriteRune(',')
					}
					i++
				}
				out.WriteRune('}')
				if idx+1 < len(rows) {
					out.WriteRune(',')
				}
			}
			out.WriteRune(']')
			result, _ = otto.ToValue(out.String())
		}

		return result
	}
}

func (a *application) getExecDBFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		if query, err := call.Argument(0).ToString(); err == nil {
			var arguments []interface{}
			for i := 1; i < len(call.ArgumentList); i++ {
				arg, _ := call.Argument(i).Export()
				arguments = append(arguments, arg)
			}
			res := a.ExecDB(query, arguments)

			if res != nil {
				lastInsertId, _ := res.LastInsertId()
				rowsAffected, _ := res.RowsAffected()
				result, _ = otto.ToValue(fmt.Sprintf("{ \"lastInsertId\" : \"%d\", \"rowsAffected\" : \"%d\"}", lastInsertId, rowsAffected))
			}
		}

		return result
	}
}

func (a *application) getEnvFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		if env, err := call.Argument(0).ToString(); err == nil {
			result, _ = otto.ToValue(os.Getenv(env))
		}

		return result
	}
}

func (a *application) getSleepFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if ms, err := call.Argument(0).ToInteger(); err == nil {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

		return otto.Value{}
	}
}

func (a *application) getDoGetFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		if aURL, err := call.Argument(0).ToString(); err == nil {
			//request parameters
			var params map[string]interface{}
			if paramsInterface, err := call.Argument(1).Export(); err == nil {
				if pz, ok := paramsInterface.(map[string]interface{}); ok {
					params = pz
				}
			}
			//request headers
			var headers map[string]interface{}
			if headersInterface, err := call.Argument(2).Export(); err == nil {
				if hz, ok := headersInterface.(map[string]interface{}); ok {
					headers = hz
				}
			}
			//request timeout
			timeout := 30
			if call.Argument(3).IsNumber() {
				if t, err := call.Argument(3).ToInteger(); err == nil {
					timeout = int(t)
				}
			}
			result, _ = otto.ToValue(a.doGet(aURL, params, headers, timeout))

		}

		return result
	}
}

func (a *application) getDoPostFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}

		if aURL, err := call.Argument(0).ToString(); err == nil {
			//request parameters
			var params map[string]interface{}
			if paramsInterface, err := call.Argument(1).Export(); err == nil {
				if pz, ok := paramsInterface.(map[string]interface{}); ok {
					params = pz
				}
			}
			//request headers
			var headers map[string]interface{}
			if headersInterface, err := call.Argument(2).Export(); err == nil {
				if hz, ok := headersInterface.(map[string]interface{}); ok {
					headers = hz
				}
			}
			//request timeout
			timeout := 30
			if call.Argument(3).IsNumber() {
				if t, err := call.Argument(3).ToInteger(); err == nil {
					timeout = int(t)
				}
			}
			result, _ = otto.ToValue(a.doPOST(aURL, params, headers, timeout))

		}

		return result
	}
}

func (a *application) getReplaceOptionsFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if chatID, err := call.Argument(0).ToString(); err == nil {
			if msgID, err := call.Argument(1).ToInteger(); err == nil {
				if optionsInterface, err := call.Argument(2).Export(); err == nil {
					if inlineOptions, ok := optionsInterface.([]map[string]interface{}); ok {
						a.replaceInlineOptions(chatID, int(msgID), inlineOptions)
					}
				}
			}
		}

		return otto.Value{}
	}
}

func (a *application) getDeleteMessageFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if chatID, err := call.Argument(0).ToString(); err == nil {
			if msgID, err := call.Argument(1).ToInteger(); err == nil {
				a.deleteMessage(chatID, int(msgID))
			}
		}

		return otto.Value{}
	}
}

func (a *application) getEditMessageFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if chatID, err := call.Argument(0).ToString(); err == nil {
			if msgID, err := call.Argument(1).ToInteger(); err == nil {
				if text, err := call.Argument(2).ToString(); err == nil {
					if optionsInterface, err := call.Argument(3).Export(); err == nil {
						if inlineOptions, ok := optionsInterface.([]map[string]interface{}); ok {
							a.editMessage(chatID, int(msgID), text, inlineOptions)
						}
					}
				}
			}
		}

		return otto.Value{}
	}
}

func (a *application) getGetFileLinkFunc() func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}
		if call.Argument(0).IsString() {
			fileID, _ := call.Argument(0).ToString()
			result, _ = otto.ToValue(a.getFileLink(fileID))
		}

		return result
	}
}

func (a *application) getSetFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			key, _ := call.Argument(0).ToString()
			if call.Argument(1).IsObject() {
				val := call.Argument(1).Object()
				a.setCacheItem(fmt.Sprintf("%s_%s", userID, key), val)
			}
		}

		return otto.Value{}
	}
}

func (a *application) getGetFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		result := otto.Value{}
		if call.Argument(0).IsString() {
			key, _ := call.Argument(0).ToString()
			result, _ = otto.ToValue(a.getCacheItem(fmt.Sprintf("%s_%s", userID, key)))
		}

		return result
	}
}

func (a *application) getDelFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			key, _ := call.Argument(0).ToString()
			a.delCacheItem(fmt.Sprintf("%s_%s", userID, key))
		}

		return otto.Value{}
	}
}

func (a *application) getPromptFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var text, attachment, targetUser string

		if call.Argument(0).IsString() {
			text, _ = call.Argument(0).ToString()
			text = strings.TrimSpace(text)
		}
		if call.Argument(1).IsString() {
			attachment, _ = call.Argument(1).ToString()
			attachment = strings.TrimSpace(attachment)
		}

		targetUser = userID
		if call.Argument(2).IsDefined() && !call.Argument(2).IsNull() {
			if tu, err := call.Argument(2).ToString(); err == nil {
				targetUser = tu
			}
		}

		id := a.promptUser(targetUser, text, attachment)

		result, _ := otto.ToValue(id)

		return result
	}
}

func (a *application) getSendFunc(userID string) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var text, attachment, targetUser string
		var options [][]string
		var inlineOptions []map[string]interface{}

		if call.Argument(0).IsString() {
			text, _ = call.Argument(0).ToString()
			text = strings.TrimSpace(text)
		}
		optionsInterface, err := call.Argument(1).Export()

		if err == nil {
			var ok bool
			if options, ok = optionsInterface.([][]string); !ok {
				inlineOptions, _ = optionsInterface.([]map[string]interface{})
			}
		}

		if call.Argument(2).IsString() {
			attachment, _ = call.Argument(2).ToString()
			attachment = strings.TrimSpace(attachment)
		}

		targetUser = userID
		if call.Argument(3).IsDefined() && !call.Argument(3).IsNull() {
			if tu, err := call.Argument(3).ToString(); err == nil {
				targetUser = tu
			}
		}

		id := a.sendMessage(targetUser, text, options, inlineOptions, attachment)

		result, _ := otto.ToValue(id)

		return result
	}
}

func (a *application) promptUser(userID string, text string, attachment string) int {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered in promptUser ", r)
		}
	}()

	attachmentFile := filepath.Join(a.attachmentsDir, attachment)
	hasAttachment := attachment != "" && FileExists(attachmentFile)

	var id int
	var err error

	if hasAttachment {
		fileType := GetFileType(attachmentFile)
		if fileType == PHOTO {
			id, err = a.tgClient.AttachPhoto(userID, attachmentFile, text, tbot.OptForceReply)
		} else if fileType == VIDEO {
			id, err = a.tgClient.AttachVideo(userID, attachmentFile, text, tbot.OptForceReply)
		} else if fileType == AUDIO {
			id, err = a.tgClient.AttachAudio(userID, attachmentFile, text, tbot.OptForceReply)
		} else {
			id, err = a.tgClient.AttachFile(userID, attachmentFile, text, tbot.OptForceReply)
		}
	} else if attachment != "" {
		fileParts := strings.Split(attachment, ":")
		if len(fileParts) == 2 {
			fileType := ParseFileType(fileParts[1])
			if fileType == PHOTO {
				id, err = a.tgClient.ForwardPhoto(userID, fileParts[0], text, tbot.OptForceReply)
			} else if fileType == VIDEO {
				id, err = a.tgClient.ForwardVideo(userID, fileParts[0], text, tbot.OptForceReply)
			} else if fileType == AUDIO {
				id, err = a.tgClient.ForwardAudio(userID, fileParts[0], text, tbot.OptForceReply)
			} else {
				id, err = a.tgClient.ForwardFile(userID, fileParts[0], text, tbot.OptForceReply)
			}
		} else {
			id, err = a.tgClient.ForwardFile(userID, attachment, text, tbot.OptForceReply)
		}
	} else if strings.TrimSpace(text) != "" {
		id, err = a.tgClient.SendText(userID, text, tbot.OptForceReply)
	} else {
		log.Warn("Ignoring empty response")
	}

	if err != nil {
		log.Error("Error prompting user ", err)
	}

	return id
}

func (a *application) sendMessage(userID string, text string, options [][]string, inlineOptions []map[string]interface{}, attachment string) int {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered in sendMessage ", r)
		}
	}()

	attachmentFile := filepath.Join(a.attachmentsDir, attachment)
	hasAttachment := attachment != "" && FileExists(attachmentFile)
	hasOptions := len(options) > 0
	hasInlineOptions := len(inlineOptions) > 0

	var id int
	var err error

	if hasAttachment {
		//file uploading
		fileType := GetFileType(attachmentFile)
		if hasOptions {
			if fileType == PHOTO {
				id, err = a.tgClient.AttachPhoto(userID, attachmentFile, text, tbot.OptReplyKeyboardMarkup(
					buildReplyOptions(options),
				))
			} else if fileType == VIDEO {
				id, err = a.tgClient.AttachVideo(userID, attachmentFile, text, tbot.OptReplyKeyboardMarkup(
					buildReplyOptions(options),
				))
			} else if fileType == AUDIO {
				id, err = a.tgClient.AttachAudio(userID, attachmentFile, text, tbot.OptReplyKeyboardMarkup(
					buildReplyOptions(options),
				))
			} else {
				id, err = a.tgClient.AttachFile(userID, attachmentFile, text, tbot.OptReplyKeyboardMarkup(
					buildReplyOptions(options),
				))
			}

		} else if hasInlineOptions {
			if fileType == PHOTO {
				id, err = a.tgClient.AttachPhoto(userID, attachmentFile, text, tbot.OptInlineKeyboardMarkup(
					buildInlineOptions(inlineOptions),
				))
			} else if fileType == VIDEO {
				id, err = a.tgClient.AttachVideo(userID, attachmentFile, text, tbot.OptInlineKeyboardMarkup(
					buildInlineOptions(inlineOptions),
				))
			} else if fileType == AUDIO {
				id, err = a.tgClient.AttachAudio(userID, attachmentFile, text, tbot.OptInlineKeyboardMarkup(
					buildInlineOptions(inlineOptions),
				))
			} else {
				id, err = a.tgClient.AttachFile(userID, attachmentFile, text, tbot.OptInlineKeyboardMarkup(
					buildInlineOptions(inlineOptions),
				))
			}
		} else {
			if fileType == PHOTO {
				id, err = a.tgClient.AttachPhoto(userID, attachmentFile, text, tbot.OptReplyKeyboardRemove)
			} else if fileType == VIDEO {
				id, err = a.tgClient.AttachVideo(userID, attachmentFile, text, tbot.OptReplyKeyboardRemove)
			} else if fileType == AUDIO {
				id, err = a.tgClient.AttachAudio(userID, attachmentFile, text, tbot.OptReplyKeyboardRemove)
			} else {
				id, err = a.tgClient.AttachFile(userID, attachmentFile, text, tbot.OptReplyKeyboardRemove)
			}
		}
	} else if attachment != "" {
		//file forwarding
		fileParts := strings.Split(attachment, ":")
		if len(fileParts) == 2 {
			//file type is specified
			fileType := ParseFileType(fileParts[1])
			if hasOptions {
				if fileType == PHOTO {
					id, err = a.tgClient.ForwardPhoto(userID, fileParts[0], text, tbot.OptReplyKeyboardMarkup(
						buildReplyOptions(options),
					))
				} else if fileType == VIDEO {
					id, err = a.tgClient.ForwardVideo(userID, fileParts[0], text, tbot.OptReplyKeyboardMarkup(
						buildReplyOptions(options),
					))
				} else if fileType == AUDIO {
					id, err = a.tgClient.ForwardAudio(userID, fileParts[0], text, tbot.OptReplyKeyboardMarkup(
						buildReplyOptions(options),
					))
				} else {
					id, err = a.tgClient.ForwardFile(userID, fileParts[0], text, tbot.OptReplyKeyboardMarkup(
						buildReplyOptions(options),
					))
				}
			} else if hasInlineOptions {
				if fileType == PHOTO {
					id, err = a.tgClient.ForwardPhoto(userID, fileParts[0], text, tbot.OptInlineKeyboardMarkup(
						buildInlineOptions(inlineOptions),
					))
				} else if fileType == VIDEO {
					id, err = a.tgClient.ForwardVideo(userID, fileParts[0], text, tbot.OptInlineKeyboardMarkup(
						buildInlineOptions(inlineOptions),
					))
				} else if fileType == AUDIO {
					id, err = a.tgClient.ForwardAudio(userID, fileParts[0], text, tbot.OptInlineKeyboardMarkup(
						buildInlineOptions(inlineOptions),
					))
				} else {
					id, err = a.tgClient.ForwardFile(userID, fileParts[0], text, tbot.OptInlineKeyboardMarkup(
						buildInlineOptions(inlineOptions),
					))
				}
			} else {
				if fileType == PHOTO {
					id, err = a.tgClient.ForwardPhoto(userID, fileParts[0], text, tbot.OptReplyKeyboardRemove)
				} else if fileType == VIDEO {
					id, err = a.tgClient.ForwardVideo(userID, fileParts[0], text, tbot.OptReplyKeyboardRemove)
				} else if fileType == AUDIO {
					id, err = a.tgClient.ForwardAudio(userID, fileParts[0], text, tbot.OptReplyKeyboardRemove)
				} else {
					id, err = a.tgClient.ForwardFile(userID, fileParts[0], text, tbot.OptReplyKeyboardRemove)
				}
			}
		} else {
			//send generic document
			if hasOptions {
				id, err = a.tgClient.ForwardFile(userID, attachment, text, tbot.OptReplyKeyboardMarkup(
					buildReplyOptions(options),
				))
			} else if hasInlineOptions {
				id, err = a.tgClient.ForwardFile(userID, attachment, text, tbot.OptInlineKeyboardMarkup(
					buildInlineOptions(inlineOptions),
				))
			} else {
				id, err = a.tgClient.ForwardFile(userID, attachment, text, tbot.OptReplyKeyboardRemove)
			}
		}
	} else if hasOptions {
		id, err = a.tgClient.SendText(
			userID,
			text,
			tbot.OptReplyKeyboardMarkup(
				buildReplyOptions(options),
			),
		)
	} else if hasInlineOptions {
		id, err = a.tgClient.SendText(
			userID,
			text,
			tbot.OptInlineKeyboardMarkup(
				buildInlineOptions(inlineOptions),
			),
		)
	} else if strings.TrimSpace(text) != "" {
		id, err = a.tgClient.SendText(userID, text, tbot.OptReplyKeyboardRemove)
	} else {
		log.Warn("Ignoring empty response")
	}

	if err != nil {
		log.Error("Error sending message ", err)
	}

	return id
}

func buildReplyOptions(replyOptions [][]string) *tbot.ReplyKeyboardMarkup {
	keyboard := make([][]tbot.KeyboardButton, len(replyOptions))
	for i := range replyOptions {
		keyboard[i] = make([]tbot.KeyboardButton, len(replyOptions[i]))
		for j := range replyOptions[i] {
			keyboard[i][j] = tbot.KeyboardButton{Text: replyOptions[i][j]}
		}
	}

	return &tbot.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		OneTimeKeyboard: true,
		ResizeKeyboard:  true,
	}

}

func buildInlineOptions(inlineOptions []map[string]interface{}) *tbot.InlineKeyboardMarkup {
	keyboard := make([][]tbot.InlineKeyboardButton, len(inlineOptions))

	for i := range inlineOptions {
		theMap := inlineOptions[i]

		vals := make([]string, 0, len(theMap))
		theReversedMap := map[string]string{}
		for key, val := range theMap {
			theValue := fmt.Sprintf("%s", val)
			theReversedMap[theValue] = key
			vals = append(vals, theValue)
		}
		sort.Strings(vals)

		j := 0
		keyboard[i] = make([]tbot.InlineKeyboardButton, len(theMap))
		for _, val := range vals {
			if isValidUrl(val) {
				button := tbot.InlineKeyboardButton{
					Text: theReversedMap[val],
					URL:  val,
				}
				keyboard[i][j] = button
			} else {
				button := tbot.InlineKeyboardButton{
					Text:         theReversedMap[val],
					CallbackData: val,
				}
				keyboard[i][j] = button
			}
			j++
		}
	}

	return &tbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}
