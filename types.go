package main

import (
	"net/url"

	"database/sql"

	"github.com/ReneKroon/ttlcache"
	"github.com/robertkrimen/otto"
	"github.com/yanzay/tbot/v2"
)

type application struct {
	tgClient       Telebot
	cache          *ttlcache.Cache
	attachmentsDir string
	token          string
	vmFactory      VmFactory
	dbClient       *sql.DB
	vmTemplate     Vm
}

type Vm interface {
	Set(name string, value interface{}) error
	Run(src interface{}) (otto.Value, error)
	Call(source string, argumentList ...interface{}) (otto.Value, error)
	Object(source string) (*otto.Object, error)
	Copy() Vm
}

type VmWrapper struct {
	vm *otto.Otto
}

func (v VmWrapper) Set(name string, value interface{}) error {
	return v.vm.Set(name, value)
}

func (v VmWrapper) Run(src interface{}) (otto.Value, error) {
	return v.vm.Run(src)
}

func (v VmWrapper) Call(source string, argumentList ...interface{}) (otto.Value, error) {
	return v.vm.Call(source, nil, argumentList...)
}

func (v VmWrapper) Object(source string) (*otto.Object, error) {
	return v.vm.Object(source)
}

func (v VmWrapper) Copy() Vm {
	return &VmWrapper{vm: v.vm.Copy()}
}

type VmFactory interface {
	GetVm() Vm
}

type VmFactoryImpl struct {
}

func (v VmFactoryImpl) GetVm() Vm {
	return &VmWrapper{vm: otto.New()}
}

type Telebot interface {
	GetFileInfo(fileID string) (*tbot.File, error)
	AnswerCallback(callbackQueryID string) error
	EditInlineMarkup(chatID string, messageID int, markup *tbot.InlineKeyboardMarkup) (int, error)
	AttachPhoto(chatID string, filename string, text string, option func(r url.Values)) (int, error)
	AttachVideo(chatID string, filename string, text string, option func(r url.Values)) (int, error)
	AttachAudio(chatID string, filename string, text string, option func(r url.Values)) (int, error)
	AttachFile(chatID string, filename string, text string, option func(r url.Values)) (int, error)
	ForwardPhoto(chatID string, fileID string, text string, option func(r url.Values)) (int, error)
	ForwardVideo(chatID string, fileID string, text string, option func(r url.Values)) (int, error)
	ForwardAudio(chatID string, fileID string, text string, option func(r url.Values)) (int, error)
	ForwardFile(chatID string, fileID string, text string, option func(r url.Values)) (int, error)
	SendText(chatID string, text string, option func(r url.Values)) (int, error)
	DeleteMsg(chatID string, messageID int) error
	EditMsg(chatID string, messageID int, text string, markup *tbot.InlineKeyboardMarkup) error
}

type TbotWrapper struct {
	*tbot.Client
}

func (t *TbotWrapper) AnswerCallback(callbackQueryID string) error {
	return t.AnswerCallbackQuery(callbackQueryID)
}

func (t *TbotWrapper) GetFileInfo(fileID string) (*tbot.File, error) {
	return t.GetFile(fileID)
}

func (t *TbotWrapper) DeleteMsg(chatID string, messageID int) error {
	return t.DeleteMessage(chatID, messageID)
}

func (t *TbotWrapper) EditMsg(chatID string, messageID int, text string, markup *tbot.InlineKeyboardMarkup) error {
	_, err := t.EditMessageText(chatID, messageID, text, tbot.OptParseModeHTML, tbot.OptInlineKeyboardMarkup(markup))
	return err
}

func (t *TbotWrapper) EditInlineMarkup(chatID string, messageID int, markup *tbot.InlineKeyboardMarkup) (int, error) {
	msg, err := t.EditMessageReplyMarkup(chatID, messageID, tbot.OptInlineKeyboardMarkup(markup))
	return msg.MessageID, err
}

func (t *TbotWrapper) AttachPhoto(chatID string, filename string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendPhotoFile(chatID, filename, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) AttachVideo(chatID string, filename string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendVideoFile(chatID, filename, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) AttachAudio(chatID string, filename string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendAudioFile(chatID, filename, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) AttachFile(chatID string, filename string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendDocumentFile(chatID, filename, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) ForwardPhoto(chatID string, fileID string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendPhoto(chatID, fileID, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) ForwardVideo(chatID string, fileID string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendVideo(chatID, fileID, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) ForwardAudio(chatID string, fileID string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendAudio(chatID, fileID, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) ForwardFile(chatID string, fileID string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendDocument(chatID, fileID, tbot.OptCaption(text), tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}

func (t *TbotWrapper) SendText(chatID string, text string, option func(r url.Values)) (int, error) {
	msg, err := t.SendMessage(chatID, text, tbot.OptParseModeHTML, option)
	return msg.MessageID, err
}
