package mongodb

// This is a value (de|en)coder for the github.com/google/uuid UUID type. For best experience, register
// mongoRegistry to mongo client instance via options, e.g.
//  clientOptions := options.Client().SetRegistry(mongoRegistry)
//
// Only BSON binary subtype 0x04 is supported.

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"reflect"
)

var (
	tID         = reflect.TypeOf(cronjob.ID{})
	uuidSubtype = byte(0x04)
)

func NewRegistry() *bsoncodec.Registry {
	registry := bson.NewRegistry()
	registry.RegisterTypeEncoder(tID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	registry.RegisterTypeDecoder(tID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))

	return registry
}

func uuidEncodeValue(_ bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tID {
		return bsoncodec.ValueEncoderError{Name: "uuidEncodeValue", Types: []reflect.Type{tID}, Received: val}
	}
	b := val.Interface().(cronjob.ID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

func uuidDecodeValue(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tID {
		return bsoncodec.ValueDecoderError{Name: "uuidDecodeValue", Types: []reflect.Type{tID}, Received: val}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		err = vr.ReadNull()
	case bson.TypeUndefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(cronjob.ID(uuid2)))
	return nil
}
