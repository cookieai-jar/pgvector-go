package pgx

import (
	"database/sql/driver"
	"fmt"

	"github.com/cookieai-jar/pgvector-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type VectorCodec struct{}

func (VectorCodec) FormatSupported(format int16) bool {
	return format == pgx.BinaryFormatCode || format == pgx.TextFormatCode
}

func (VectorCodec) PreferredFormat() int16 {
	return pgx.BinaryFormatCode
}

func (VectorCodec) PlanEncode(m *pgtype.Map, oid uint32, format int16, value any) pgtype.EncodePlan {
	_, ok := value.(pgvector.Vector)
	if !ok {
		return nil
	}

	switch format {
	case pgx.BinaryFormatCode:
		return encodePlanVectorCodecBinary{}
	case pgx.TextFormatCode:
		return encodePlanVectorCodecText{}
	}

	return nil
}

type encodePlanVectorCodecBinary struct{}

func (encodePlanVectorCodecBinary) Encode(value any, buf []byte) (newBuf []byte, err error) {
	v := value.(pgvector.Vector)
	return v.EncodeBinary(buf)
}

type encodePlanVectorCodecText struct{}

func (encodePlanVectorCodecText) Encode(value any, buf []byte) (newBuf []byte, err error) {
	v := value.(pgvector.Vector)
	// use String() for now to avoid adding another method to Vector
	return append(buf, v.String()...), nil
}

func (VectorCodec) PlanScan(m *pgtype.Map, oid uint32, format int16, target any) pgtype.ScanPlan {
	_, ok := target.(*pgvector.Vector)
	if !ok {
		return nil
	}

	switch format {
	case pgx.BinaryFormatCode:
		return scanPlanVectorCodecBinary{}
	case pgx.TextFormatCode:
		return scanPlanVectorCodecText{}
	}

	return nil
}

type scanPlanVectorCodecBinary struct{}

func (scanPlanVectorCodecBinary) Scan(src []byte, dst any) error {
	v := (dst).(*pgvector.Vector)
	return v.DecodeBinary(src)
}

type scanPlanVectorCodecText struct{}

func (scanPlanVectorCodecText) Scan(src []byte, dst any) error {
	v := (dst).(*pgvector.Vector)
	return v.Scan(src)
}

func (c VectorCodec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return c.DecodeValue(m, oid, format, src)
}

func (c VectorCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var vec pgvector.Vector
	scanPlan := c.PlanScan(m, oid, format, &vec)
	if scanPlan == nil {
		return nil, fmt.Errorf("Unable to decode vector type")
	}

	err := scanPlan.Scan(src, &vec)
	if err != nil {
		return nil, err
	}

	return vec, nil
}
