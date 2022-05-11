package jsonapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/ebladrocher/keypass/pass"
	"github.com/pkg/errors"
)

var (
	sep = "/"
)

func (api *API) respondMessage(ctx context.Context, msgBytes []byte) error {
	var message messageType
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "не удалось десериализовать сообщение JSON")
	}

	switch message.Type {
	case "query":
		return api.respondQuery(msgBytes)
	case "queryHost":
		return api.respondHostQuery(msgBytes)
	case "getLogin":
		return api.respondGetLogin(ctx, msgBytes)
	case "create":
		return api.respondCreateEntry(ctx, msgBytes)
	default:
		return fmt.Errorf("Сообщение неизвестного типа %s", message.Type)
	}
}

func (api *API) respondQuery(msgBytes []byte) error {
	var message queryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "не удалось десериализовать сообщение JSON")
	}

	l, err := api.Store.List()
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}

	choices := make([]string, 0, 10)
	reQuery := fmt.Sprintf(".*%s.*", regexSafeLower(message.Query))
	if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
		return errors.Wrapf(err, "не удалось добавить результаты поиска")
	}

	return sendSerializedJSONMessage(choices, api.Writer)
}

func (api *API) respondHostQuery(msgBytes []byte) error {
	var message queryHostMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "не удалось десериализовать сообщение JSON")
	}

	l, err := api.Store.List()
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}
	choices := make([]string, 0, 10)

	for !isPublicSuffix(message.Host) {
		// only query for paths and files in the store fully matching the hostname.
		reQuery := fmt.Sprintf("(^|.*/)%s($|/.*)", regexSafeLower(message.Host))
		if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
			return errors.Wrapf(err, "не удалось добавить результаты поиска")
		}
		if len(choices) > 0 {
			break
		} else {
			message.Host = strings.SplitN(message.Host, ".", 2)[1]
		}
	}

	return sendSerializedJSONMessage(choices, api.Writer)
}

func (api *API) respondGetLogin(ctx context.Context, msgBytes []byte) error {
	var message getLoginMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "не удалось десериализовать сообщение JSON")
	}

	sec, err := api.Store.Get(message.Entry)
	if err != nil {
		return errors.Wrapf(err, "не удалось получить секрет")
	}

	return sendSerializedJSONMessage(loginResponse{
		Username: api.getUsername(message.Entry),
		Password: string(sec),
	}, api.Writer)
}

func (api *API) getUsername(name string) string {
	// if no meta-data was found return the name of the secret itself
	// as the username, e.g. providers/amazon.com/foobar -> foobar
	if strings.Contains(name, sep) {
		return path.Base(name)
	}

	return ""
}

func searchAndAppendChoices(reQuery string, list []string, choices *[]string) error {
	re, err := regexp.Compile(reQuery)
	if err != nil {
		return errors.Wrapf(err, "не удалось скомпилировать регулярное выражение '%s': %s", reQuery, err)
	}

	for _, value := range list {
		if re.MatchString(strings.ToLower(value)) {
			*choices = append(*choices, value)
		}
	}
	return nil
}

func sendSerializedJSONMessage(message interface{}, w io.Writer) error {
	// we can't use json.NewEncoder(w).Encode because we need to send the final
	// message length before the actul JSON
	serialized, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := writeMessageLength(serialized, w); err != nil {
		return err
	}

	var msgBuf bytes.Buffer
	count, err := msgBuf.Write(serialized)
	if err != nil {
		return err
	}
	if count != len(serialized) {
		return fmt.Errorf("сообщение не полностью написано ")
	}

	wcount, err := msgBuf.WriteTo(w)
	if wcount != int64(len(serialized)) {
		return fmt.Errorf("сообщение не полностью написано ")
	}
	return err
}

func writeMessageLength(msg []byte, w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, uint32(len(msg)))
}

func (api *API) respondCreateEntry(ctx context.Context, msgBytes []byte) error {
	var message createEntryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "не удалось десериализовать сообщение JSON")
	}

	tmp, err := api.Store.Exists(message.Name)
	if err != nil {
		return fmt.Errorf("ERROR %s", err)
	}
	if tmp {
		return fmt.Errorf("секрет %s уже существует", message.Name)
	}

	if message.Generate {
		str := pass.GeneratePassword(message.PasswordLength, message.UseSymbols)
		if err != nil {
			return fmt.Errorf("ERROR %s", err)
		}
		message.Password = string(str)
	}

	if err := api.Store.SetConfirm(message.Name, []byte(message.Password), api.confirmRecipients); err != nil {
		return errors.Wrapf(err, "failed to store secret")
	}

	return sendSerializedJSONMessage(loginResponse{
		Username: message.Login,
		Password: message.Password,
	}, api.Writer)
}
