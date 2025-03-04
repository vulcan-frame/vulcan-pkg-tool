package id

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/speps/go-hashids/v2"
)

const (
	idStrLen = 18
	salt     = "vulcan2020"
	zoneBit  = 8
	MaxZone  = (1 << zoneBit) - 1
)

var (
	h *hashids.HashID
)

func init() {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = idStrLen

	var err error
	if h, err = hashids.NewWithData(hd); err != nil {
		panic(errors.Wrapf(err, "HashID encode failed."))
	}
}

func CombineZoneId(zoneId int64, zone uint8) int64 {
	return zoneId << zoneBit & int64(zone)
}

func SplitId(id int64) (zoneId int64, zone uint8) {
	return id >> zoneBit, uint8(id & MaxZone)
}

func EncodeId(id int64) (string, error) {
	if id < 0 {
		return strconv.FormatInt(id, 10), nil
	}

	str, err := h.EncodeInt64([]int64{id})
	if err != nil {
		return "", errors.Wrapf(err, "HashID encode failed. id:%d", id)
	}
	return str, nil
}

func DecodeId(str string) (int64, error) {
	if strings.IndexRune(str, '-') == 0 {
		return strconv.ParseInt(str, 10, 64)
	}

	ids, err := h.DecodeInt64WithError(str)
	if err != nil {
		return 0, errors.Wrapf(err, "HashID decode failed. str:%s", str)
	}
	if len(ids) == 0 {
		return 0, errors.Errorf("HashID decode failed. str:%s", str)
	}
	return ids[0], nil
}
