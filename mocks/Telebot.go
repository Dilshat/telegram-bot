// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import tbot "github.com/yanzay/tbot/v2"
import url "net/url"

// Telebot is an autogenerated mock type for the Telebot type
type Telebot struct {
	mock.Mock
}

// AnswerCallback provides a mock function with given fields: callbackQueryID
func (_m *Telebot) AnswerCallback(callbackQueryID string) error {
	ret := _m.Called(callbackQueryID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(callbackQueryID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AttachAudio provides a mock function with given fields: chatID, filename, text, option
func (_m *Telebot) AttachAudio(chatID string, filename string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, filename, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, filename, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, filename, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AttachFile provides a mock function with given fields: chatID, filename, text, option
func (_m *Telebot) AttachFile(chatID string, filename string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, filename, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, filename, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, filename, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AttachPhoto provides a mock function with given fields: chatID, filename, text, option
func (_m *Telebot) AttachPhoto(chatID string, filename string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, filename, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, filename, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, filename, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AttachVideo provides a mock function with given fields: chatID, filename, text, option
func (_m *Telebot) AttachVideo(chatID string, filename string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, filename, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, filename, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, filename, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EditInlineMarkup provides a mock function with given fields: chatID, messageID, markup
func (_m *Telebot) EditInlineMarkup(chatID string, messageID int, markup *tbot.InlineKeyboardMarkup) (int, error) {
	ret := _m.Called(chatID, messageID, markup)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, int, *tbot.InlineKeyboardMarkup) int); ok {
		r0 = rf(chatID, messageID, markup)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int, *tbot.InlineKeyboardMarkup) error); ok {
		r1 = rf(chatID, messageID, markup)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ForwardAudio provides a mock function with given fields: chatID, fileID, text, option
func (_m *Telebot) ForwardAudio(chatID string, fileID string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, fileID, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, fileID, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, fileID, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ForwardFile provides a mock function with given fields: chatID, fileID, text, option
func (_m *Telebot) ForwardFile(chatID string, fileID string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, fileID, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, fileID, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, fileID, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ForwardPhoto provides a mock function with given fields: chatID, fileID, text, option
func (_m *Telebot) ForwardPhoto(chatID string, fileID string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, fileID, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, fileID, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, fileID, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ForwardVideo provides a mock function with given fields: chatID, fileID, text, option
func (_m *Telebot) ForwardVideo(chatID string, fileID string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, fileID, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, fileID, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, fileID, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFileInfo provides a mock function with given fields: fileID
func (_m *Telebot) GetFileInfo(fileID string) (*tbot.File, error) {
	ret := _m.Called(fileID)

	var r0 *tbot.File
	if rf, ok := ret.Get(0).(func(string) *tbot.File); ok {
		r0 = rf(fileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tbot.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(fileID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendText provides a mock function with given fields: chatID, text, option
func (_m *Telebot) SendText(chatID string, text string, option func(url.Values)) (int, error) {
	ret := _m.Called(chatID, text, option)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, func(url.Values)) int); ok {
		r0 = rf(chatID, text, option)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, func(url.Values)) error); ok {
		r1 = rf(chatID, text, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
